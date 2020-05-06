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
kcfi deploy [ -c config.yaml ]
```

### Example - Codefresh onprem installation
```
kcfi init codefresh
```
It created `codefresh` directory with config.yaml and other files

- Edit `codefresh/config.yaml` - set global.appUrl and other parameters  
- Set docker registry credentials  
- Set tls certifcates (optional)  

Deploy Codefresh
```
kcfi deploy -c codefresh/config.yaml
```

