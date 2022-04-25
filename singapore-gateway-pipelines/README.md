# Singapore Gateway Pipelines

## About

Singapore Gateway Pipelines provides a Tekton Pipeline which will do the following - 
1. Create an OSD cluster with standard configuration
2. Configure OSD cluster with authentication
    1. Creates an cluster-admin user bot account with credentials provided in [secrets_template.yaml](prereqs/secrets_template.yaml)
    2. Configures GitHub IDP access to team provided
3. Configure OSD cluster
    1. Adds ImageContentSourcePolicy for downstream images
    2. Updates cluster global pull-secret to ensure images can be pulled
    3. Deploys OpenShift GitOps Operator for ArgoCD
4. Apply MultiClusterEngine
    1. Applies MultiClusterEngine by [argo application](resources/multiclusterengine/mce-argo.yaml)
    2. Application files are stored [here](../argo/operators/multicluster-engine/)
    3. MCE configuration is exactly [here](../argo/operators/multicluster-engine/multiclusterengine.yaml). Currently ManagedServiceAccount is enabled

Miro diagrams detailing how the Pipeline functions and how to deploy it can be found [here](https://miro.com/app/board/uXjVO-ggYrA=/)

## Prereqs

1. OCP Cluster 4.8+
2. Create a [GitHub Oauth App](https://docs.github.com/en/developers/apps/building-oauth-apps/creating-an-oauth-app)
    1. Be sure to note the `GitHub Client ID` and `GitHub Client Secret`
    2. We will not have the `Homepage URL` or `Authorization callback URL` yet. We will update these at the end in the [post running pipeline](#post-running-pipeline) steps
3. Fill out [secrets_template.yaml](prereqs/secrets_template.yaml)
   1. Information on how to retrieve values provided in the template

## Deploy

To deploy run the following commands - 

```bash
oc apply -f singapore-gateway-pipelines/prereqs/secrets_template.yaml --namespace <NAMESPACE>
oc apply -f singapore-gateway-pipelines/ --namespace <NAMESPACE>
```

## Running Pipeline

To run the Pipeline, this can be done through the UI simply - 
1. Navigate to Pipelines in OpenShift console
    1. Sign into OCP console
    2. On side nav bar, dropdown `Pipelines`
    3. Click Pipelines and ensure the Project selected is the namespace the Pipeline was deployed too
2. Click `deploy-singapore-gateway` Pipeline
3. Click `Actions` then `Start`
4. In the interface which opens, fill in the required fields -
    1. Param - `clusterName` - this will be the name of the OSD cluster to create. If this cluster exists already, the pipeline will skip the creation step and use the existing cluster. 
    2. Param -`clusterLifetime` - this is the length of time in hours which the cluster will be alive for. The default is 168h.
    3. Workspace - `shared-workpace` - Select `PersistentVolumeClaim`. In the dropdown that follows, select a PVC. A simple PVC can be created if one does not exist.
4. Click `Start`


## Post Running Pipeline

After the Pipeline has succeeded, the cluster has configured as expected. The final step is to ensure GitHub is configured to provide IDP to the cluster.

Complete the following to ensure GitHub can provide IDP - 

1. Sign into OCM -
    1. Connect to Red Hat VPN
    2. Navigate to [Staging OCM](https://qaprodauth.cloud.redhat.com/)
    3. Sign in
2. Click `OpenShift` in the side nav bar
3. Click on the cluster we created by running the pipeline
    1. If there are many clusters in this view, a cluster can be searched for, or filtered to show only OSD clusters.
4. Click `Access control` tab
5. Click `Identity providers` tab
6. Copy the Github `Auth callback URL` by clicking `Copy URL to clipboard` in the `Github` identity provider row
7. In a separate tab navigate to [GitHub](https://github.com) and sign in
8. Navigate to account settings
9. Go to `Developer settings`
10. Click `OAuth apps`
11. Click on the GitHub Ouath app we created in the prereqs step
12. Paste the OSD clusters GitHub `Auth callback URL` in the `Authorization callback URL` field
13. Add the `Homepage URL`
    1. This will be the OpenShift console URL. This value can be found in the previous tab by clicking `Open console` and copying the URL which opens.
14. It may take ~10 minutes for IDP changes to take effect on the cluster after the Pipeline has succeeded

