# Power Dynamic Resource Allocation Driver

This repository contains a Power Architecture dynamic resource allocation driver for use with the *Dynamic Resource Allocation (DRA)* feature of Kubernetes. The driver aligns with [KEP-4381: Dynamic Resource Allocation with Structured Parameters](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/4381-dra-structured-parameters/README.md) and assigns at most one device to one pod on the system.

The driver facilitates access to:

1. *Nest Accelerator unit (NX) compression and decompression co-processors*: NX is a non-core part of the Power processor, which provides a DEFLATE compliant (RFC1950, 1951, 1952) compression accelerator. This feature has a device driver in userspace `/dev/crypto/nx-gzip` which is shared among logical-partitions on the same systems, and by proxy containers in the logical partition. 

As additional features are made available, the driver is to be expanded.

This project is based on [dra-example-driver](https://github.com/kubernetes-sigs/dra-example-driver). This project is licensed under the Apache 2.0 License.

## Pre-Requisites

On an OpenShift Container Platform cluster, one must enable the `DynamicResourceAllocation` `FeatureGate`. To setup, go into the Cluster UI, `Administration` -> `CustomResourceDefinitions` -> `FeatureGate` -> `Instances` -> `cluster` add `spec.featureSet: TechPreviewNoUpgrade`

You must have Power10 or higher logical partitions that are part of your cluster. It will not allocate `/dev/crypto/nx-gzip` access. The system must have licensed the Power system feature `active_mem_expansion_capable`.

You should also be using an OpenShift Container Platform 4.19+.

## Kind cluster

To use the kind cluster in the power arch machine, run the script to create the kind cluster with single worker node.

``` shell
ARCH=arm64 make dev-setup
```

## Install

To install the code, use:

``` shell
helm install \
  --create-namespace \
  --namespace power-dra-driver \
  power-dra-driver \
  deployments/helm/power-dra-driver
```

## Upgrade

To install the code, use:

``` shell
helm upgrade \
  -n power-dra-driver \
  --namespace power-dra-driver \
  power-dra-driver \
  deployments/helm/power-dra-driver
```

## Uninstall

To uninstall the code, use:

``` shell
helm uninstall \
    power-dra-driver \
    -n power-dra-driver
```

## Delete the kind cluster

To delete the kind cluster created , run the script to delete it.

``` shell
ARCH=arm64 make dev-teardown
```

## Trouble

To get the logs for the driver, use:

``` shell
oc logs -n power-dra-driver -l app.kubernetes.io/name=power-dra-driver
```

## License

All source files must include a Copyright and License header. The SPDX license header is 
preferred because it can be easily scanned.

If you would like to see the detailed LICENSE click [here](LICENSE).

``` text
/*
 * Copyright 2025 - IBM Corporation. All rights reserved
 * SPDX-License-Identifier: Apache-2.0
 */
```

## References

1. [Kubernetes: Dynamic Resource Allocation](https://kubernetes.io/docs/concepts/scheduling-eviction/dynamic-resource-allocation/)
2. [Kubernetes: Dynamic Resource Allocation example](https://github.com/kubernetes-sigs/dra-example-driver/blob/main/README.md)
3. [OpenShift Feature Gate: DynamicResourceAllocation](https://docs.openshift.com/container-platform/4.17/nodes/clusters/nodes-cluster-enabling-features.html)

> Enables a new API for requesting and sharing resources between pods and containers. This is an internal feature that most users do not need to interact with. (DynamicResourceAllocation)

OpenShift Docs are specific to a use-case

This code repository is based on [kubernetes-sigs/dra-example-driver](https://github.com/kubernetes-sigs/dra-example-driver) and extends APIs used in the example driver.

# Support

> Is this a Red Hat or IBM supported solution?

This solution is used to demonstrate accessing the co-processors part of the IBM Power systems. It is a demonstration at this time.