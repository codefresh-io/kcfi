# Installation Warnings

## Default passwords
We recommend set passwords for persistent components (mongo, postgres, redis, rabbit ), otherwise default known password will be used  

Add or uncomment the following strings in config.yaml under `global:`:  
```yaml
# mongodb passwords for root and cfuser
global:
  #mongodbRootUser: root
  mongodbRootPassword: <rootPassword>
  
  #mongodbUsername: cfuser
  mongodbPassword: <cfuserPassword>

  #postgresUser: postgres
  postgresPassword: <thePassword>

  # rabbit password
  rabbitmqUsername: 
  rabbitmqPassword: <thePassword>

  # redis password
  redisPassword: <thePassword>
```
