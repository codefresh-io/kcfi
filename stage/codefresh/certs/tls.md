### Web SSL Certificates for Codefresh installer
installer configures ingress tls patameters accorfing to  "tls" key in values.yaml

```yaml
# default values
tls:
  selfSigned: false
  cert: certs/ssl.crt
  key: certs/private.key
```

if ssl.selfSigned=false (default) installer validates and uses values of ssl.cert and ssl.key.
Certifaicate and key files should exist in the specified location.
Otherwise if ssl.selfSigned=true it generates selfsigned certificates with CN=<global.appUrl>