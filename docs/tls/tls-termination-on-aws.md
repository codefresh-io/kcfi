# TLS termination on AWS

To use either a certificate from a third party issuer that was uploaded to IAM or a certificate [requested](https://docs.aws.amazon.com/acm/latest/userguide/gs-acm-request-public.html) within AWS Certificate Manager:
- copy a certificate ARN;
- set the `tls.selfSigned: true` in the Codefresh's init config - __config.yaml__;
- deploy a new installation;
- update ingress service

```sh
kubectl edit service cf-ingress-controller
```

and add the following annotations:

```yaml
metadata:
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-backend-protocol: http
    service.beta.kubernetes.io/aws-load-balancer-ssl-cert: < CERTIFICATE ARN >
    service.beta.kubernetes.io/aws-load-balancer-ssl-ports: "443"
spec:
  ports:
  - name: http
    nodePort: 30908
    port: 80
    protocol: TCP
    targetPort: 80
  - name: https
    nodePort: 31088
    port: 443
    protocol: TCP
    targetPort: 80
```

Both HTTP and HTTPS target port should be set to **80**.

> ToDo
Add automation for this to be able to point a certificate ARN in init config.