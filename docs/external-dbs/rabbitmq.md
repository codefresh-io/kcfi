# Configuring external RabbitMQ service

To use an external RabbitMQ service instead of local helm chart, add the following values to the __config.yaml__:

```yaml
rabbitmq:
  enabled: false
  rabbitmqUsername: <RABBITMQ USER>
  rabbitmqPassword: <RABBITMQ PASSWORD> 

global:
  rabbitmqHostname: <RABBITMQ HOST>
```