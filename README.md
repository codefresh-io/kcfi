# `kcfi` - Codefresh Installer for Kubernetes  

### Download
https://github.com/codefresh-io/kcfi/releases

### Usage
Create configuration directory
```
kcfi init <product> [-d /path/to/stage-dir]
```
Edit configuration in config.yaml and deploy to Kubernetes
```
kcfi deploy [ -c config.yaml ] [ --kube-context <kube-context-name> ] [ --atomic ] [ --debug ] [ helm upgrade parameters ]
```

### Example - Codefresh onprem installation
```
kcfi init codefresh
```
It creates `codefresh` directory with config.yaml and other files

- Edit `codefresh/config.yaml` - set global.appUrl and other parameters  
- Set docker registry credentials - obtain sa.json from Codefesh or set your private registry address and credentials  
- Set tls certifcates (optional) - set tls.selfSigned=false put ssl.crt and private.key into certs/ directory 
- For Openshift uncomment `include: ["values/openshift.yaml"]` in config.yaml

Deploy Codefresh
```
kcfi deploy -c codefresh/config.yaml [ --kube-context <kube-context-name> ] [ --atomic ] [ --debug ] [ helm upgrade parameters ]
```

### Separate database infrastructure chart
By default Codefresh installation includes persistent services (mongo, postgres, redis, rabbitmq). You can optionally install them as separate chart by setting dbinfra.enabled=true in config.yaml :
```yaml
dbinfra:
  enabled: true
 #storageClass: nfs-with-backup
```
You can also specify storageClass and other chart values.  
This is preferable option for installation on Openshift  

### Registering external docker nodes
Edit [codefresh-stage-dir/addons/external-nodes/config.yaml](stage/addons/external-nodes/config.yaml) to set node addresses  
```yaml
metadata:
  kind: helmChart
  installer:
    type: helm
    helm:
      chart: external-nodes
      release: cf-external-nodes

kubernetes:
  namespace: codefresh
  #context: 
  #kubeconfig:

runtimeEnvironments:
  - name: default-nodes
    extends: system/default
    cluster: codefresh
    protocol: http
    #alternateLoggerConf: external-nodes
    nodes:
    - address: "172.31.128.95"
      name: node-1
      port: 4243
    - address: "172.31.128.96"
      name: node-2
      port: 4243
```

Deploy nodes configuration by
```
kcfi deploy -c codefresh/addons/external-nodes/config.yaml
```

### Openshift Routes
Set `host` value in `codefresh-stage-dir/addons/openshift-routes/config.yaml` and deploy:  
```
kcfi deploy -c codefresh/addons/openshift-routes/config.yaml --disable-openapi-validation
```

### Example - cf-k8s-agent installation
1. Run `kcfi init k8s-agent`. A staging directory will be created named `k8s-agent` with config.yaml and other files.
2. Edit `k8s-agent/config.yaml`.
3. Run `/kcfi deploy -c k8s-agent/config.yaml -n your_namespace` standing on proper kube context.

### Uploading images to private registry in air-gapped environment
(without access to public docker hub and codefresh-enterprise registry)
```
Push whole release with images list defined in config file or by --image-list parameter:
   kcfi images push [--images-list <images-list-file>] [-c|--config /path/to/config.yaml] [options]

Push single image
  kcfi images push [-c|--config /path/to/config.yaml] [options] repo/image:tag [repo/image:tag] ...

Usage:
  kcfi images [flags]

Aliases:
  images, image, private-registry, docker

Flags:
      --codefresh-registry-secret string   file with Codefresh registry secret (sa.json)
  -c, --config string                      config file
  -h, --help                               help for images
      --images-list string                 file with list of images to push
      --password string                    registry password
      --registry string                    registry address
      --user string                        registry username
```

### Additional docs
see in [docs](./docs) folder 