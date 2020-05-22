# Managing Codefresh backups

Codefresh on-premise backups can be automated by installing a specific service named "Backup Manager". It is based on [this](https://github.com/stefanprodan/mgob) open source project and can run scheduled backups with retention, S3 & SFTP upload, notifications, instrumentation with Prometheus and more.

### Configuring and Installing Backup Manager

1. Initialize the backup manager staging directory:
```
kcfi init backup-manager [-d /path/to/stage-dir]
```
2. If you need to change backup manager configuration, please edit the `config.yaml`, which have been generated in the staging directory. 

It is recommended to leave the mongo uri value blank - it will be taken automatically from the Codefresh release installed in your cluster.

3. Select a correct kube context, where you have Codefresh on-premise installed and deploy backup-manager with the following command:

```
kcfi deploy -c /path/to/stage-dir/config.yaml -n your_namespace
```
### On-demand backup
```
kubectl port-forward cf-backup-manager-0 8090
curl -X POST http://localhost:8090/backup/cfBackupPlan
```

### Restore
```
kubectl exec -it cf-backup-manager-0 bash
mongorestore --gzip --archive=/storage/cfBackupPlan/cfBackupPlan-1590156430.gz --uri mongodb://root:password@mongodb:27017 --drop
```