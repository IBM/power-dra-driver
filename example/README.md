

1. Create namespace `oc new-project ex-dra-driver`
2. Install Scenario - `kustomize build examples/<scenario> | oc apply -f -`
3. Uninstall Scenario - `kustomize build examples/<scenario> | oc delete -f -`


The scenarios are:
1. `two-pods-two-distinct-nx-gzip` - Two pods, one container each, Each container asking for access to the nx-gzip

These examples are inspired by https://github.com/kubernetes-sigs/dra-example-driver/tree/main/demo