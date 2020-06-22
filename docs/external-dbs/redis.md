# Configure external Redis

Codefresh requires two Redis database:

- the main - `cf-redis`, to store sessions, cache, etc;
- `cf-store`, to store triggers;
  
The first one can be replaced by external Redis service, the `cf-store` unfortunately no.

>ToDo
update __hermes/redis__ chart to be able to use an external Redis as `cf-store`.

To configure Codefresh to use an external Redis service, add the following parameters to __config.yaml__:

```yaml
redis:
  enabled: false
  redisPassword: <MY REDIS PASS>

global:
  redisUrl: <MY REDIS HOST>
  runtimeRedisHost: <MY REDIS HOST>
  runtimeRedisPassword: <MY REDIS PASS>
  runtimeRedisDb: 2
```
