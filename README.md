# Power Dynamic Resource Allocation Driver
This repository contains a Power Architecture resource driver for use with the Dynamic Resource Allocation (DRA) feature of Kubernetes.

The driver facilitates access to:
1. Nest Accelerator unit (NX) are non-core components of the POWER processor chip including compression and encryption co-processors. One feature is nx-gzip, which is a DEFLATE compliant (RFC1950, 1951, 1952) compression accelerator in NX. This feature is a device driver in userspace `/dev/crypto/nx-gzip` which is shared among many LPARs and many containers. 

As additional features are made available, the driver will be expanded.

This project is licensed under the Apache 2.0 License.

## Pre-Requisites

This feature uses DynamicResourceAllocation. To setup, go into the Cluster UI, `Administration` -> `CustomResourceDefinitions` -> `FeatureGate` -> `Instances` -> `cluster` add `spec.featureSet: TechPreviewNoUpgrade`

You must have Power10+ nodes to take advantage of the power-dra-driver. It will not allocate nx-gzip access if you are not on a Power10, or if the nx-gzip is some how disabled on the Operating System.

You should also be using an OpenShift Container Platform 4.19+.

## Install

```
helm upgrade \
  --create-namespace \
  --namespace power-dra-driver \
  power-dra-driver \
  deployments/helm/power-dra-driver
```

## Uninstall

```
helm uninstall power-dra-driver -n power-dra-driver
```

## References
1. [Kubernetes: Dynamic Resource Allocation](https://kubernetes.io/docs/concepts/scheduling-eviction/dynamic-resource-allocation/)
2. [Kubernetes: Dynamic Resource Allocation example](https://github.com/kubernetes-sigs/dra-example-driver/blob/main/README.md)
3. [OpenShift Feature Gate: DynamicResourceAllocation](https://docs.openshift.com/container-platform/4.17/nodes/clusters/nodes-cluster-enabling-features.html)
> Enables a new API for requesting and sharing resources between pods and containers. This is an internal feature that most users do not need to interact with. (DynamicResourceAllocation)
OpenShift Docs are specific to a use-case

This code repository is based on [kubernetes-sigs/dra-example-driver](https://github.com/kubernetes-sigs/dra-example-driver) and extends apis used in the example driver.