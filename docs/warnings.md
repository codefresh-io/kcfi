# Installation Warnings

## Default passwords
We recommend set passwords for persistent components (mongo, postgres, redis, rabbit ), otherwise default known password will be used  

Add or uncomment the following strings in config.yaml:  
```yaml
# mongodb passwords for root and cfuser
mongodb:
  mongodbRootPassword: <rootPassword>
  mongodbPassword: <cfuserPassword>

# postgres password
postgresql:
  postgresPassword: <thePassword>

# redis password
redis:
  redisPassword: <thePassword>

# rabbit password
rabbitmq:
  rabbitmqPassword: <thePassword>
```
