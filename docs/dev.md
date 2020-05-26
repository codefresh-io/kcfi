# kcfi - Developer guide

### Abstract
We are using "helm.sh/helm/v3" and "k8s.io/cli-runtime" libraries to interact with helm charts and kubernetes manifests

### Adding new product (chart)
- create product stage directory in [stage folder](stage/). Create config.yaml and optionally other related files and folders
```yaml
# config.yaml example
metadata: 
  kind: new-product
  installer: 
    # type: 
    #   "operator" - apply crd definition, operator controller should be running in the cluster or kcfi code should run it
    #   "helm" - install/upgrade helm chart from client
    type: helm

    # operator parameters (if type=operator)
    operator:
      image: codefresh/cf-onprem-operator:test
      serviceAccountName:
      skipCRD: false
    # helm chart parameters (if type=helm)
    helm:
      chart: new-product
      #repoUrl: http://charts.codefresh.io/stable
      #version:

# Configure docker image pull secret and images list
images:
  codefreshRegistrySa: sa.json
  # usePrivateRegistry: false
  # privateRegistry:
  #   address:
  #   username:
  #   password:
  lists:
  - images/images-list

### Optional Helm Chart values and other configurations
aaa: va
ccc:
  c1: aaa
  c2: bbb
```
- If the product is deployed as an embeded helm chart, copy chart as a directory with Chart.yaml or as an archived chart (like chartname-1.2.3.tgz) to [charts folder](charts/).

- If there is a custom logic in addition to helm install/upgrade, add code to handle new metadata.kind to [apply.go](pkg/action/apply.go)

Helm chart and stage files are embeded into the binary - see [build.sh](hack/build.sh) and its result in [pkg/embeded](pkg/embeded)

### helpers:
- [helm.go](pkg/action/helm.go)
- [helpers.go](pkg/action/helpers.go)
- [loader.go](pkg/charts/loader.go)

### Codefresh Onprem
Deploy logic is in [codefresh.go](pkg/apply/codefresh.go)
- validation of config.yaml
- docker image pull secret generation
- dbInfra chart if dbinfra.enabled=true
- tls certificates 
- changing values depending installing/updating
- Can deploy to Kubernetes as operator or as embeded helm chart depending on `metadata.installer.type`

[Operator chart](charts/codefresh-operator)
[Stage](stage/codefresh)

### TODO
* Codefresh-onprem - add storage handling and validator like we have in old repo 
* Automatic images-list generation depending on release name (in cooperation with cd pipeline)













