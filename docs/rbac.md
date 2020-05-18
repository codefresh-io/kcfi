# RBAC for Codefresh

Codefresh installer should be run with Kubernetes rbac role which allows to create all kinds of object in single namespace.  
In case when creation of service account or role is not allowed by the corporate policy, ask Kubernetes administrator to create the role+serviceAccount+binding as below. Users with the codefresh-app role below cannot create other roles and roleBindings

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: codefresh-app
  namespace: codefresh
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: codefresh-app
  namespace: codefresh
rules:
- apiGroups:
  - ""
  - apps
  - codefresh.io
  - autoscaling
  - extensions
  - batch
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - networking.k8s.io
  - route.openshift.io
  - policy
  resources:
  - routes
  - ingresses
  - poddisruptionbudgets
  verbs:
  - '*'

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    app: codefresh
  name: codefresh-app-binding
  namespace: codefresh
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: codefresh-app
subjects:
- kind: ServiceAccount
  name: codefresh-app
```

