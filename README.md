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

Deploy Codefresh
```
kcfi deploy -c codefresh/config.yaml [ --kube-context <kube-context-name> ] [ --atomic ] [ --debug ] [ helm upgrade parameters ]
```

### Example - cf-k8s-agent installation
1. Run `kcfi init k8s-agent`. A staging directory will be created named `k8s-agent` with config.yaml and other files.
2. Edit `k8s-agent/config.yaml`.
3. Run `/kcfi deploy -c k8s-agent/config.yaml -n your_namespace` standing on proper kube context.