# Configuring external Postgres database

it is possible to configure Codefresh to work with user-provided Postgres database service, in case if default one provided automatically within Codefresh installation is not applicable for the user. This document describes how to do that.

#### Configuration steps

All the configuration comes down to putting a set of correct values into your Codefresh configuration file `config.yaml`, which is present in `your/stage-dir/codefresh` directory. During the installation, Codefresh will run a seed job, using the values described in the below steps:

1. Specify a user name `global.postgresSeedJob.user` and password `global.postgresSeedJob.password` for a seed job. This must be a privileged user allowed to create databases and roles. It will be used only by the seed job to create the needed database and a user.
2. Specify a user name `global.postgresUser` and password `global.postgresPassword` to be used by Codefresh installation. A user with the name and password will be created by the seed job and granted with required privileges to access the created database.
3. Specify a database name `global.postgresDatabase` to be created by the seed job and used by Codefresh installation.
4. Specify `global.postgresHostname` and optionally `global.postgresPort` (`5432` is a default value).
5. Disable the postgres subchart installation with the `postgresql.enabled: false` value, because it is not needed in this case.

Below is an example of a relevant piece of configuration YAML would look like:

```yaml
global:
  postgresSeedJob:
    user: postgres
    password: zDyGp79XyZEqLq7V
  postgresUser: cf_user
  postgresPassword: fJTFJMGV7sg5E4Bj
  postgresDatabase: codefresh
  postgresHostname: my-postgres.ccjog7pqzunf.us-west-2.rds.amazonaws.com
  postgresPort: 5432

postgresql:
  enabled: false
```

#### Running the seed job manually

In case if you'd prefer running the seed job manually, you could do it by using a script present in `your/stage-dir/codefresh/addons/seed-scripts` directory named `postgres-seed.sh`. The script takes the following set of variables you need to have set before running it:

```
POSTGRES_SEED_USER="postgres"
POSTGRES_SEED_PASSWORD="zDyGp79XyZEqLq7V"
POSTGRES_USER="cf_user"
POSTGRES_PASSWORD="fJTFJMGV7sg5E4Bj"
POSTGRES_DATABASE="codefresh"
POSTGRES_HOST="my-postgres.ccjog7pqzunf.us-west-2.rds.amazonaws.com"
POSTGRES_PORT="5432"
```
The variables have the same meaning as the configuratoin values described in the above section of this document.

However you **still need to specify a set of values** in the Codefresh config file as described in the section above, but with the whole **`postgresSeedJob` section ommitted**, like this:

```yaml
global:
  postgresUser: cf_user
  postgresPassword: fJTFJMGV7sg5E4Bj
  postgresDatabase: codefresh
  postgresHostname: my-postgres.ccjog7pqzunf.us-west-2.rds.amazonaws.com
  postgresPort: 5432

postgresql:
  enabled: false
```