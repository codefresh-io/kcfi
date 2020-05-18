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
      chart: codefresh-1.0.99.tgz
      #repo
      #version:

# Configure docker image pull secret
docker:
  codefreshRegistrySa: sa.json
  # usePrivateRegistry: false
  # privateRegistry:
  #   address:
  #   username:
  #   password:
  #   configFile:

### Optional Helm Chart values and other configurations
aaa: va
ccc:
  c1: aaa
  c2: bbb
```
- If the product is deployed as a helm chart copy chart as a directory with Chart.yaml or as an archived chart (like chartname-1.2.3.tgz) to [charts folder](charts/)

- Add code to handle new metadata.kind to [apply.go](pkg/action/apply.go)

Helm chart and stage files are embeded into the binary - see [build.sh](hack/build.sh) and its result in [pkg/embeded](pkg/embeded)  

### helpers:
- [helm.go](pkg/action/helm.go)
- [templates.go](pkg/action/templates.go)
- [loader.go](pkg/charts/loader.go)

### Codefresh Onprem
Deploy logic is in [codefresh.go](pkg/apply/codefresh.go)
- validation of config.yaml
- docker image pull secret generation
- tls certificates 
- changing values depending installing/updating
- Can deploy to Kubernetes as operator or as embeded helm chart depending on `metadata.installer.type`

[Operator chart](charts/codefresh-operator)
[Stage](stage/codefresh)

### TODO
* Implement `kcfi private-registry push|push-release`

* support adding generic helm chart: developers just put chart into [charts](charts/) and stage/<new-product>/config.yaml with `metadata.installer.helm: {chart: chartname, version: 1.2.3}` and installer will apply the chart with values without extra code

* Codefresh-onprem - add storage handling and validator like we have in old repo 

* 













