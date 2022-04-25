# HyperShift Pipelines

## About

This repo is a contains a set of Pipelines and their associated Tasks to deploy ACM, enable HyperShift, and deploy a HyperShiftDeployment.
Their are 2 main Pipelines contained in this repo, one which functions in the Upstream, and another in the Downstream.

## Prereqs

The following prereqs are required - 

1. A ClusterPool in the same namespace of the Pipelines. See [clusterpool.yaml](../hypershift-pipelines-prereqs/clusterpool.yaml) for an example. It is recommended to have a size of at least 2. HyperShift requires OCP 4.10.7 in order to run a HostedCluster.
2. The following secrets must be defined in the same namespace of the Pipelines. See [secrets.yaml](../hypershift-pipelines-prereqs/secrets.yaml)
3. An OCP Cluster (4.8+ recommended) with the following Operators Installed - 
    * OpenShift Pipelines
    * Advanced Cluster Management (2.4+ recommended)


## How to deploy

To deploy the Pipelines run - 

```
$ oc apply -f hypershift-pipelines -n <NAMESPACE>
```

## Cleanup

To cleanup HyperShift resources and the clusters after deploying via either the upstream or downstream pipelines, run the [cleanup pipeline](pipeline_cleanup.yaml) - `cleanup-acm-and-hypershift-deployment`.

This pipeline takes in 1 paramter - the name of the clusterclaim used in either the upstream or downstream Pipelines and completes the following operations - 

1. Checks out existing hub cluster and logs into it
2. Deletes all HyperShiftDeployments on the hub
3. Detaches the imported spoke cluster
4. Deletes the MultiClusterHub custom resource
5. Deletes both the spoke and hub clusterclaims
