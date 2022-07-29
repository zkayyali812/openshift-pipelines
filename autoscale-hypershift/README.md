# Autoscale HyperShift 

## About

Autoscale Hypershift consists of 2 cronjob which can be applied to your cluster, to ensure your HyperShiftDeployments are able to scale-up/down on a schedule.

A scale up job, enables autoscaling with a minimum of 2 replicas, and a maximum of 5.

A scale down job, disables autoscaling and sets replicas to 1.

The image for this can be built, and tested with OpenShift Pipelines.

## Deploying via templates

### Prereqs

The only prereq is to create and fill an `options.env` file. This can be done by running - 

```
$ cd autoscale-hypershift
$ make options.env
```

Add the following to this file 

```
NAMESPACE=<namespace-to-deploy-templates-into>
```

### Deploy cronjobs

To deploy as is, we can clone this repository and run the command below.

```
$ cd autoscale-hypershift
$ make roles cronjobs
```

### Manually trigger scaling

To scale up/down once, by running just the job, the commands below can be run

```
$ cd autoscale-hypershift
$ make roles
$ make scale-up/scale-down
```

### Running locally (Development)

The Go program can be ran directly, targetting your current cluster context by running the commands below - 

NOTE: Running locally uses the current RBAC privelges of your current context.

```
$ cd autoscale-hypershift
$ make scale-up-local/scale-down-local
```

## Deploying via Policy

Autoscale HyperShift can also be deployed via Advanced Cluster Management Policies.

The following [sample](templates/policy-example.yaml) can be applied to ensure autoscale-hypershift is enabled via policy.

## Pipelines

OpenShift Pipelines are included to show examples of building, and building and testing the image via OpenShift Pipelines.

### Build Pipeline

This [Pipeline](pipeline/pipeline.yaml) is a simple build pipeline which will compile, build, and push the autoscale-hypershift image. This image will be tagged per the parameter given in the Pipeline.

If pushing to a private registry, ensure the `pipelines` service account is linked to an imagePullSecret authorized access to this registry.

### Test and Build Pipeline

This [Pipeline](pipeline/test_pipeline.yaml) is a more comlex pipeline which will build the image, checkout a cluster, deploy a HyperShiftDeployment, then run the scale-up and scale-down jobs, and ensure their success. If successful, it will delete its resources and retag the image, to publish it. This image will be tagged per the parameter given in the Pipeline.

If pushing to a private registry, ensure the `pipelines` service account is linked to an imagePullSecret authorized access to this registry.


