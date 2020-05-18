# Operator CRD

In case when corporate security rules do not allow crd creation for a client which runs `kcfi`, codefresh crd and its serviceAccout+Role should be created separately.  
Ask cluster administrator to create the crd and rbac and then set `skipCRD=true` and `serviceAccocuntName=codefresh-app` in config.yaml 

* Codefresh CRD
```yaml
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: codefreshes.codefresh.io
  labels:
    app: cf-onprem-operator
spec:
  group: codefresh.io
  names:
    kind: Codefresh
    listKind: CodefreshList
    plural: codefreshes
    singular: codefresh
  scope: Namespaced
  subresources:
    status: {}
  versions:
  - name: v1alpha1
    served: true
    storage: true
```

* ServiceAccout+Role+RoleBinding: see [rbac.md](./rbac.md)

* In config.yaml set `metadata.operator.skipCRD: true` and `metadata.operator.serviceAccountName: codefresh-app` to run the operator with preexisting crd and serviceAccount:
```yaml
# config.yaml
metadata:
  kind: codefresh
  installer:
    # type: 
    #   "operator" - apply codefresh crd definition
    #   "helm" - install/upgrade helm chart from client
    type: operator
    operator:
      #dockerRegistry: gcr.io/codefresh-enterprise
      #image: codefresh/cf-onprem-operator
      #imageTag: 1.0.99
      serviceAccountName: codefresh-app
      skipCRD: true
...
```