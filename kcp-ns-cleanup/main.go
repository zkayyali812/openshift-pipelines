package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	hyperdeployv1alpha1 "github.com/stolostron/hypershift-deployment-controller/api/v1alpha1"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var hypershiftDeploymentRes = schema.GroupVersionResource{Group: "cluster.open-cluster-management.io", Version: "v1alpha1", Resource: "hypershiftdeployments"}
var secretRes = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
var namespaceRes = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
var customResourceDefinitionRes = schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}
var resourcesToCheck = []schema.GroupVersionResource{
	// {Group: "", Version: "v1", Resource: "bindings"},
	{Group: "", Version: "v1", Resource: "configmaps"},
	{Group: "", Version: "v1", Resource: "endpoints"},
	// {Group: "", Version: "v1", Resource: "events"},
	{Group: "", Version: "v1", Resource: "limitranges"},
	{Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
	{Group: "", Version: "v1", Resource: "pods"},
	{Group: "", Version: "v1", Resource: "podtemplates"},
	{Group: "", Version: "v1", Resource: "replicationcontrollers"},
	{Group: "", Version: "v1", Resource: "resourcequotas"},
	{Group: "", Version: "v1", Resource: "secrets"},
	{Group: "", Version: "v1", Resource: "serviceaccounts"},
	{Group: "", Version: "v1", Resource: "services"},
	{Group: "apps", Version: "v1", Resource: "controllerrevisions"},
	{Group: "apps", Version: "v1", Resource: "daemonsets"},
	{Group: "apps", Version: "v1", Resource: "deployments"},
	{Group: "apps", Version: "v1", Resource: "replicasets"},
	{Group: "apps", Version: "v1", Resource: "statefulsets"},
}

// Simple error function
func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
}

func main() {
	// Determine what action to take ScaleDown or ScaleUp

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

	unstructuredHyperDeploys, err := client.Resource(hypershiftDeploymentRes).List(context.TODO(), metav1.ListOptions{})
	checkError(err)

	for _, unstructuredHyperdeploy := range unstructuredHyperDeploys.Items {
		fmt.Println("== Checking HypershiftDeployment: ", unstructuredHyperdeploy.GetName(), " ==")
		hd := &hyperdeployv1alpha1.HypershiftDeployment{}
		convertUnstructuredToHypershiftDeployment(unstructuredHyperdeploy, hd)

		hostedClusterKubeConfigSecretName := fmt.Sprintf("%s-admin-kubeconfig", hd.Name)
		hostedClusterKubeConfigSecretUnstructured, err := client.Resource(secretRes).Namespace(hd.Namespace).Get(context.TODO(), hostedClusterKubeConfigSecretName, metav1.GetOptions{})
		checkError(err)
		hostedClusterKubeConfigSecret := &v1.Secret{}
		convertUnstructuredToSecret(*hostedClusterKubeConfigSecretUnstructured, hostedClusterKubeConfigSecret)
		hostedClusterKubeConfig := string(hostedClusterKubeConfigSecret.Data["kubeconfig"])
		cleanupKCPNamespaces(hd, hostedClusterKubeConfig)
		fmt.Println(" ")
	}
}

func cleanupKCPNamespaces(hd *hyperdeployv1alpha1.HypershiftDeployment, kubeconfig string) {
	// If the file doesn't exist, create it, or append to the file
	kubeConfigFilename := fmt.Sprintf("%s-kubeconfig", hd.Name)
	managedClusterName := hd.Spec.HostedClusterSpec.InfraID

	f, err := os.OpenFile(kubeConfigFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	checkError(err)
	defer f.Close()
	f.Write([]byte(kubeconfig))
	defer os.Remove(kubeConfigFilename)

	var config *rest.Config
	fmt.Printf("Connecting with hypershiftdeployment %s kubeconfig\n", hd.Name)
	config, err = clientcmd.BuildConfigFromFlags("", kubeConfigFilename)
	checkError(err)

	client, err := dynamic.NewForConfig(config)
	checkError(err)

	unstructuredNamespaces, err := client.Resource(namespaceRes).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("internal.workload.kcp.dev/cluster=%s", managedClusterName),
	})
	checkError(err)

	for _, unstructuredNamespace := range unstructuredNamespaces.Items {
		// hd := &hyperdeployv1alpha1.HypershiftDeployment{}
		unstructuredCRDs, err := client.Resource(customResourceDefinitionRes).List(context.TODO(), metav1.ListOptions{})
		checkError(err)
		checkIfResourcesExistInHostedCluster(client, unstructuredNamespace.GetName(), unstructuredCRDs.Items)
	}
}

func checkIfResourcesExistInHostedCluster(client dynamic.Interface, namespace string, unstructuredCRDs []unstructured.Unstructured) {
	resourcesFound := false
	fmt.Printf("-> Checking for core resources in namespace: %s\n", namespace)
	for _, resourceToCheck := range resourcesToCheck {
		unstructuredResources, err := client.Resource(resourceToCheck).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
		checkError(err)
		if len(unstructuredResources.Items) > 0 {
			remainingResources := filterOutAllowedResources(unstructuredResources.Items, resourceToCheck.Resource)
			if len(remainingResources) > 0 {
				fmt.Printf("--> Found %d %s\n", len(remainingResources), resourceToCheck.Resource)
				for _, resourceName := range remainingResources {
					fmt.Println("---> - ", resourceName.GetName())
				}
				resourcesFound = true
			}
		}
	}

	fmt.Printf("-> Checking for non corev1 resources in namespace: %s\n", namespace)
	for _, unstructuredCRD := range unstructuredCRDs {
		crd := &apixv1.CustomResourceDefinition{}
		convertUnstructuredToCRD(unstructuredCRD, crd)

		if crd.Spec.Scope != "Namespaced" {
			// only checks namespaced resources
			continue
		}
		// Check if namespaced scoped resources exist in hosted cluster

		for _, version := range crd.Spec.Versions {
			crdGVK := schema.GroupVersionResource{Group: crd.Spec.Group, Version: version.Name, Resource: crd.Spec.Names.Plural}
			unstructuredResources, err := client.Resource(crdGVK).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
			checkError(err)
			if len(unstructuredResources.Items) > 0 {
				fmt.Printf("--> Found %d %s\n", len(unstructuredResources.Items), crd.Spec.Names.Plural)
				for _, resourceName := range unstructuredResources.Items {
					fmt.Println("---> - ", resourceName.GetName())
				}
				resourcesFound = true
			}
		}
	}

	if !resourcesFound {
		fmt.Printf("-> DELETING NAMESPACE %s<-\n", namespace)
		err := client.Resource(namespaceRes).Delete(context.Background(), namespace, metav1.DeleteOptions{})
		checkError(err)
	}

}

type AllowList struct {
	Resources map[string][]string
}

func filterOutAllowedResources(resources []unstructured.Unstructured, resourceName string) []unstructured.Unstructured {
	yamlFile, err := ioutil.ReadFile("allowlist.yaml")
	checkError(err)
	allowList := &AllowList{}
	err = yaml.Unmarshal(yamlFile, allowList)
	checkError(err)

	_, ok := allowList.Resources[resourceName]
	if !ok {
		// No allowed resources exist for this resource
		return resources
	}

	allowedResources := allowList.Resources[resourceName]

	disallowedResources := []unstructured.Unstructured{}

	for _, resource := range resources {
		if !contains(allowedResources, resource.GetName(), resourceName) {
			disallowedResources = append(disallowedResources, resource)
		}
	}

	return disallowedResources

}

func contains(resourceList []string, resourceName, resourceType string) bool {
	for _, resource := range resourceList {
		if resourceType == "secrets" && strings.Contains(resourceName, resource) {
			return true
		} else if resource == resourceName {
			return true
		}
	}
	return false
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

func convertUnstructuredToSecret(unstructuredObj unstructured.Unstructured, hd *v1.Secret) {
	byteHyperDeploy, err := yaml.Marshal(unstructuredObj.Object)
	checkError(err)
	err = yaml.Unmarshal(byteHyperDeploy, hd)
	checkError(err)
}

func convertUnstructuredToCRD(unstructuredObj unstructured.Unstructured, crd *apixv1.CustomResourceDefinition) {
	byteHyperDeploy, err := yaml.Marshal(unstructuredObj.Object)
	checkError(err)
	err = yaml.Unmarshal(byteHyperDeploy, crd)
	checkError(err)
}
