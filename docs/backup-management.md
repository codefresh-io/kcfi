# Managing Codefresh backups

Codefresh on-premise backups can be automated by installing a specific service as an addon to your Codefresh on-premise installation. It is based on [this](https://github.com/stefanprodan/mgob) open source project and can run scheduled backups with retention, S3 & SFTP upload, notifications, instrumentation with Prometheus and more.

### Configuring and Installing Backup Manager

Backup manager is installed as an addon and therefore it needs an existing Codefresh on-premise installation. Before installing it, please make sure you have select a proper kube config pointing to the cluster, where you have Codefresh installed on.

To configure backup manager, please go to the staging directory of your Codefresh installation and find a specific config file: `your-CF-stage-dir/addons/backup-manager/config.yaml`.

There you will find a few configuration parameters, which you might want to change:

* `storage`- storage class, storage size and read modes for persistent volumes to store backups locally within your cluster
* Backup plan configuration parameters under `jobConfigs.cfBackupPlan`:
    * `target.uri` - target mongo URI. It is recommended to leave the mongo uri value blank - it will be taken automatically from the Codefresh release installed in your cluster
    * `scheduler` - here you can specify cron expression for your backups schedule, backups retention and timeout values

For more advanced backup plan settings, like specifiying various remote cloud-based storages for your backups, configuring notifications and other, please refer to [this](https://github.com/stefanprodan/mgob#configure) page 

To **deploy backup manager** service, please select a correct kube context, where you have Codefresh on-premise installed and deploy backup-manager with the following command:

```
kcfi deploy -c `your-CF-stage-dir/addons/backup-manager/config.yaml` -n your_namespace
```

### On-demand backup
```
kubectl port-forward cf-backup-manager-0 8090
curl -X POST http://localhost:8090/backup/cfBackupPlan
```

### Restore
```
kubectl exec -it cf-backup-manager-0 bash
mongorestore --gzip --archive=/storage/cfBackupPlan/backup-arhive-name.gz --uri mongodb://root:password@mongodb:27017 --drop
```