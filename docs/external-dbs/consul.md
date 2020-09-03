# Configuring external Consul service

## Limitations

At the moment Codefresh supports only deprecated Consul API (image __consul:1.0.0__), does not support connection via HTTPS and any authentication.
The Consul host have to expose port `8500`.
In general, we don't recommend to take the Consul service outside the cluster.

## config.yaml

To configure Codefresh using external Consul service, add the following values to the __config.yaml__:

```yaml
global:
  consulHost: <MY CONSUL HOST>

consul:
  enabled: false
```
