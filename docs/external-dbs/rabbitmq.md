# Configuring external RabbitMQ service

Codefresh recommends to use Bitnami RabbitMQ [chart](https://github.com/bitnami/charts/tree/master/bitnami/rabbitmq) as a RabbitMQ service.

To use an external RabbitMQ service instead of local helm chart, add the following values to the __config.yaml__:

```yaml
rabbitmq:
  enabled: false
  rabbitmqUsername: <RABBITMQ USER>
  rabbitmqPassword: <RABBITMQ PASSWORD> 

global:
  rabbitmqHostname: <RABBITMQ HOST>
```