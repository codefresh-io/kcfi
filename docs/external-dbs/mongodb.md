# Configuring external MongoDB

Codefresh recommends to use Bitnami MongoDB [chart](https://github.com/bitnami/charts/tree/master/bitnami/mongodb) as a Mongo database.

To configure Codefresh on-premises to use external Mongo service one needs to provide the following values:

- **mongo connection string** - `mongoURI`. This string will be used by all of the services to communicate with mongo. Codefresh will automatically create and add a user with "ReadWrite" permissions to all of the created databases with the username and password from the URI. Optionally, automatic user addition can be disabled - `mongoSkipUserCreation`, in order to use already existing user. In such a case the existing user must have **ReadWrite** permissions to all of newly created databases
Codefresh does not support [DNS Seedlist Connection Format](https://docs.mongodb.com/manual/reference/connection-string/#connections-dns-seedlist) at the moment, use the [Standard Connection Format](https://docs.mongodb.com/manual/reference/connection-string/#connections-standard-connection-string-format) instead.
- mongo **root user** name and **password** - `mongodbRootUser`, `mongodbRootPassword`. The privileged user will be used by Codefresh only during installation for seed jobs and for automatic user addition. After installation, credentials from the provided mongo URI will be used.  Mongo root user must have permissions to create users.

> ToDo
Even if `mongoSkipUserCreation` set to `true` and `mongoURI` contains root credentials, the `mongodbRootUser` and `mongodbRootPassword` should be set anyway because Mongo seed job expects them.
We should fix it

Here is an example of all the related values:

```yaml
global:
  mongodbRootUser: my-mongo-admin-user
  mongodbRootPassword: yeqTeVwqVa9qDqebq
  mongoURI: mongodb://someuser:mTiqweAsdw@my-mongo-cluster-shard-00-00-vziq1.mongodb.net:27017/?ssl=true
  mongoSkipUserCreation: true
  mongoDeploy: false   # disables deployment of internal mongo service

mongo:
  enabled: false
 ```
