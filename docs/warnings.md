# Installation Warnings
Installer warns if some values are not set:  

## Default passwords
We recommend set passwords for persistent components (mongo, postgres, redis, rabbit ), otherwise default known password will be used. 

Add or uncomment the following strings in config.yaml under `global:`:  
```yaml

global:
  
  #mongodbRootUser: root
  mongodbRootPassword: <rootPassword>

  # Password for codefresh app user (username = "cfuser")
  mongodbPassword: <cfuserPassword>

  #postgresUser: postgres
  postgresPassword:  <thePassword>

  # rabbit password
  rabbitmqPassword: <thePassword>

  # redis password
  redisPassword: <thePassword>
```

**Note**: To change the passwords on already installed instance:
* connect to relevant service with client using `kubectl port-forward ... ` or `kubectl exec ... ` 
* change the password (refer product guide)
* set the values above and launch `kcfi deploy ...`
