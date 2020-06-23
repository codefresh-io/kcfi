# Configuring AWS Load Balancers

By default Codefresh deploys the nginx-controller and [Classic Loab Balancer](https://docs.aws.amazon.com/eks/latest/userguide/load-balancing.html) as a controller service.

## NLB

To use a **Network Load Balancer** - deploy a regular Codefresh installation, add to the `cf-ingress-controller` controller service the `service.beta.kubernetes.io/aws-load-balancer-type: nlb` annotation.
This annotation will create a new one Load Balancer - Network Load Balancer, which you should use in Codefresh UI DNS record.
Update the DNS record according to the new service. 

## ALB

To use the **Application Load Balancer** the [ALB Ingress Controller](https://docs.aws.amazon.com/eks/latest/userguide/alb-ingress.html) should be deployed to the cluster.

To support ALB:

-  disable Nginx controller in the Codefresh init config file - __config.yaml__:

```yaml
ingress:
  enabled: false
```

- [deploy](https://docs.aws.amazon.com/eks/latest/userguide/alb-ingress.html) the ALB controller;
- create a new **ingress** resource:

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  annotations:
    alb.ingress.kubernetes.io/listen-ports: '[{"HTTP": 80}, {"HTTPS":443}]'
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/target-type: ip
    kubernetes.io/ingress.class: alb
    meta.helm.sh/release-name: cf
    meta.helm.sh/release-namespace: codefresh
  labels:
    app: cf-codefresh
    release: cf
  name: cf-codefresh-ingress
  namespace: codefresh
spec:
  backend:
    serviceName: cf-cfui
    servicePort: 80
  rules:
  - host: nosovets.cf-op.com
    http:
      paths:
      - backend:
          serviceName: cf-cfapi
          servicePort: 80
        path: /api/*
      - backend:
          serviceName: cf-cfapi
          servicePort: 80
        path: /ws/*
      - backend:
          serviceName: cf-cfui
          servicePort: 80
        path: /
```

## SSL termination

When a **Classic Load Balancer** is used, and some Codefresh features are turned on, for example `OfflineLogging`, they use websocket to connect with Codefresh API and require secure TCP (SSL) protocol for Load Balancer listener instead of HTTPS.
To support this, update the existing configuration:

- update the `cf-ingress-controller` service with new annotations:

```yaml
metadata:
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-backend-protocol: tcp
    service.beta.kubernetes.io/aws-load-balancer-ssl-ports: "443"
```

- update your AWS Load Balancer listener for port 443 from HTTPS protocol to SSL.