// author: github.com/jnpacker
package main

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	hypershiftv1alpha1 "github.com/openshift/hypershift/api/v1alpha1"
	hyperdeployv1alpha1 "github.com/stolostron/hypershift-deployment-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const AutoscaleHypershiftSA = true

// Simple error function
func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}

func scaleHyperShiftDeploymentsUpdate(client dynamic.Interface, hd *hyperdeployv1alpha1.HypershiftDeployment, action string) string {
	if action == "scaleup" {
		// Enable AutoScaling on all NodePools
		for _, nodePool := range hd.Spec.NodePools {
			if nodePool.Spec.AutoScaling == nil {
				nodePool.Spec.AutoScaling = &hypershiftv1alpha1.NodePoolAutoScaling{}
			}
			nodePool.Spec.AutoScaling = &hypershiftv1alpha1.NodePoolAutoScaling{
				Min: 2,
				Max: 5,
			}
			// nodePool.Spec.Replicas = nil
			nodePool.Spec.NodeCount = nil
		}

	} else if action == "scaledown" {
		// Disable AutoScaling and ensure 1 replica
		replicas := int32(1)
		for _, nodePool := range hd.Spec.NodePools {
			nodePool.Spec.AutoScaling = nil
			// nodePool.Spec.Replicas = &replicas
			nodePool.Spec.NodeCount = &replicas
		}
	}

	labels := hd.GetLabels()
	labels["autoscale-hypershift-currentaction"] = action
	hd.SetLabels(labels)

	// convert the runtime.Object to unstructured.Unstructured to use the dynamic client
	unstructuredHD := &unstructured.Unstructured{}
	convertHypershiftDeploymentToUnstructured(hd, unstructuredHD)

	updated, err := client.Resource(hypershiftDeploymentRes).Namespace(hd.Namespace).Update(context.TODO(), unstructuredHD, metav1.UpdateOptions{})
	checkError(err)
	return updated.GetLabels()["autoscale-hypershift-currentaction"]
}

// Used to create events for hypershiftdeployment scaling actions
func fireEvent(client dynamic.Interface, hd *hyperdeployv1alpha1.HypershiftDeployment, eventName string, message string, reason string, eType string) {
	unstructuredEvent, err := client.Resource(eventRes).Namespace(hd.Namespace).Get(context.TODO(), fmt.Sprintf("%s-%s", eventName, hd.Name), metav1.GetOptions{})
	event := &corev1.Event{}
	convertUnstructuredToEvent(unstructuredEvent, event)

	if event != nil && event.Series == nil {
		event.Series = new(corev1.EventSeries)
		event.Series.Count = 1
		event.Series.LastObservedTime = metav1.NowMicro()
	}
	if err != nil {
		fmt.Println("  |-> Event not found")
		event = new(corev1.Event)
		event.TypeMeta.Kind = "Event"
		event.TypeMeta.APIVersion = "v1"
		event.Name = fmt.Sprintf("%s-%s", eventName, hd.Name)
		event.Namespace = hd.Namespace
		event.EventTime = metav1.NowMicro()
		event.Action = reason
	}
	if event.Series != nil {
		event.Series.Count = event.Series.Count + 1
		event.Series.LastObservedTime = metav1.NowMicro()
	}
	event.Type = eType
	event.Reason = reason
	event.Message = message
	event.ReportingController = "autoscale-hypershift-cronjob"
	event.ReportingInstance = "autoscale-hypershift-cronjob"
	event.Related = &corev1.ObjectReference{
		Kind:            hd.Kind,
		Namespace:       hd.Namespace,
		Name:            hd.Name,
		UID:             hd.UID,
		ResourceVersion: hd.ResourceVersion,
		APIVersion:      fmt.Sprintf("%s/%s", hypershiftDeploymentRes.Group, hypershiftDeploymentRes.Version),
	}
	event.InvolvedObject = corev1.ObjectReference{
		Kind:            hd.Kind,
		Namespace:       hd.Namespace,
		Name:            hd.Name,
		UID:             hd.UID,
		ResourceVersion: hd.ResourceVersion,
		APIVersion:      fmt.Sprintf("%s/%s", hypershiftDeploymentRes.Group, hypershiftDeploymentRes.Version),
	}
	updatedUnstructuredEvent := &unstructured.Unstructured{}
	convertEventToUnstructured(event, updatedUnstructuredEvent)
	if err != nil {
		_, err := client.Resource(eventRes).Namespace(hd.Namespace).Create(context.TODO(), updatedUnstructuredEvent, metav1.CreateOptions{})
		fmt.Println("  -> Create a new event " + eventName + " for cluster " + hd.Namespace + "/" + hd.Name)
		checkError(err)
	} else {
		_, err := client.Resource(eventRes).Namespace(hd.Namespace).Update(context.TODO(), updatedUnstructuredEvent, metav1.UpdateOptions{})
		fmt.Println("  -> Update existing event "+eventName+", event count", event.Series.Count)
		checkError(err)
	}
}

var hypershiftDeploymentRes = schema.GroupVersionResource{Group: "cluster.open-cluster-management.io", Version: "v1alpha1", Resource: "hypershiftdeployments"}
var eventRes = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "events"}

func main() {
	// Determine what action to take ScaleDown or ScaleUp
	var TakeAction = strings.ToLower(os.Getenv("TAKE_ACTION"))
	var OptIn = os.Getenv("OPT_IN")
	if TakeAction == "" || (TakeAction != "scaledown" && TakeAction != "scaleup") {
		panic("Environment variable TAKE_ACTION missing: " + TakeAction)
	}
	homePath := os.Getenv("HOME") // Used to look for .kube/config

	var config *rest.Config
	if _, err := os.Stat(homePath + "/.kube/config"); !os.IsNotExist(err) {
		fmt.Println("Connecting with local kubeconfig")
		config, err = clientcmd.BuildConfigFromFlags("", userConfig())
		checkError(err)

	} else {
		fmt.Println("Connecting using In Cluster Config")
		config, err = rest.InClusterConfig()
		checkError(err)
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	podNamespace := os.Getenv("POD_NAMESPACE")
	unstructuredHyperDeploys, err := client.Resource(hypershiftDeploymentRes).Namespace(podNamespace).List(context.TODO(), metav1.ListOptions{})
	checkError(err)

	for _, unstructuredHyperdeploy := range unstructuredHyperDeploys.Items {
		hd := &hyperdeployv1alpha1.HypershiftDeployment{}
		convertUnstructuredToHypershiftDeployment(unstructuredHyperdeploy, hd)

		if (OptIn == "true" && hd.Labels["autoscale-hypershift"] == "true") || (OptIn != "true" && hd.Labels["autoscale-hypershift"] != "skip") {
			takeAction(client, hd, TakeAction, AutoscaleHypershiftSA)
		} else {
			fmt.Println("Skip: " + hd.Name + "  (currently " + string(hd.Labels["autoscale-hypershift"]) + ")")
			fireEvent(client, hd, "autoscale-hypershift", "Skipping cluster "+hd.Name, "skipAction", "Normal")
		}

	}
}

func takeAction(client dynamic.Interface, hd *hyperdeployv1alpha1.HypershiftDeployment, takeAction string, AutoscaleHypershiftSA bool) {
	if hd.Labels["autoscale-hypershift-currentaction"] != takeAction {
		fmt.Printf("Taking Action: %s on hd: %s\n", takeAction, hd.Name)

		newState := scaleHyperShiftDeploymentsUpdate(client, hd, takeAction)

		// Check the new state and report a response
		if newState == takeAction {
			fmt.Println("  ✓")
			if AutoscaleHypershiftSA {
				fireEvent(client, hd, "autoscale-hypershift", "The hypershiftdeployment "+hd.Name+" is taking action: "+takeAction, takeAction, "Normal")
			}
		} else {
			fmt.Println("  X")
			if AutoscaleHypershiftSA {
				fireEvent(client, hd, "autoscale-hypershift", "The cluster "+hd.Name+" did not set state to "+takeAction, "failedScaleDown", "Warning")
			}
		}
	} else {
		fmt.Println("Skip: " + hd.Name + "  (currently " + string(hd.Labels["autoscale-hypershift-currentaction"]) + ")")
		if AutoscaleHypershiftSA {
			fireEvent(client, hd, "autoscale-hypershift", "Skipping cluster "+hd.Name+", requested state "+takeAction+" equals current state "+string(hd.Labels["autoscale-hypershift-currentaction"]), "skipScaleDown", "Normal")
		}
	}
}

func userConfig() string {
	usr, err := user.Current()
	checkError(err)
	return filepath.Join(usr.HomeDir, ".kube", "config")
}

func convertUnstructuredToHypershiftDeployment(unstructuredObj unstructured.Unstructured, hd *hyperdeployv1alpha1.HypershiftDeployment) {
	byteHyperDeploy, err := yaml.Marshal(unstructuredObj.Object)
	checkError(err)
	err = yaml.Unmarshal(byteHyperDeploy, hd)
	checkError(err)
}

func convertHypershiftDeploymentToUnstructured(hd *hyperdeployv1alpha1.HypershiftDeployment, unstructuredObj *unstructured.Unstructured) {
	bytesHD, err := yaml.Marshal(hd)
	checkError(err)
	err = yaml.Unmarshal(bytesHD, unstructuredObj)
	checkError(err)
}

func convertUnstructuredToEvent(unstructuredEvent *unstructured.Unstructured, event *corev1.Event) {
	if unstructuredEvent == nil || unstructuredEvent.Object == nil {
		return
	}
	bytesEvent, err := yaml.Marshal(unstructuredEvent.Object)
	checkError(err)
	err = yaml.Unmarshal(bytesEvent, event)
	checkError(err)
}

func convertEventToUnstructured(event *corev1.Event, updatedUnstructuredEvent *unstructured.Unstructured) {
	bytesEvent, err := yaml.Marshal(event)
	checkError(err)
	err = yaml.Unmarshal(bytesEvent, updatedUnstructuredEvent)
	checkError(err)
}
