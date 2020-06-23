# Configure external Redis

Codefresh recommends to use Bitnami Redis [chart](https://github.com/bitnami/charts/tree/master/bitnami/redis) as a Redis store.

## Limitations

Codefresh does not support secure connection to Redis (TLS) and AUTH username extension.

## Configuration

Codefresh requires two Redis database:

- the main - `cf-redis`, to store sessions, cache, etc;
- `cf-store`, to store triggers;
  
Only the first one can be replaced by external Redis service,
the `cf-store` DB is used as a local storage for triggers and should run along with the installation.

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

Where `redis*` - are for the main Redis storage, and `runtimeRedis*` - for storage is used to store pipeline logs in case of `OfflineLogging` feature is turned on.
It's usually the same host.
