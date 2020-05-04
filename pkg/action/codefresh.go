/*
Copyright The Codefresh Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package action

import (
    "fmt"
	"path"
	"path/filepath"
	"os"
	"io/ioutil"

	"github.com/pkg/errors"
    "github.com/stretchr/objx"
    
	helm "helm.sh/helm/v3/pkg/action"
//	"helm.sh/helm/v3/pkg/postrender"
	"helm.sh/helm/v3/pkg/storage/driver"

	"k8s.io/cli-runtime/pkg/resource"
)

// GetDockerRegistryVars - calculater docker registry vals
func (o *CfApply) GetDockerRegistryVars () (map[string]interface{}, error) {
	
	var registryAddress, registryUsername, registryPassword string
	var err error
	valsX := objx.New(o.vals)
	usePrivateRegistry := valsX.Get(keyDockerUsePrivateRegistry).Bool(false)
	if !usePrivateRegistry {
		// using Codefresh Enterprise registry
		registryAddress = "gcr.io"
		registryUsername = "_json_key"
		cfRegistrySaVal := valsX.Get(keyDockerCodefreshRegistrySa).Str("sa.json")
		cfRegistrySaPath := path.Join(filepath.Dir(o.ConfigFile), cfRegistrySaVal)
		registryPasswordB, err := ioutil.ReadFile(cfRegistrySaPath)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("cannot read %s", cfRegistrySaPath))
		}
		registryPassword = string(registryPasswordB)
	} else {
		registryAddress = valsX.Get(keyDockerprivateRegistryAddress).String()
		registryUsername = valsX.Get(keyDockerprivateRegistryUsername).String()
		registryPassword = valsX.Get(keyDockerprivateRegistryPassword).String()
		if len(registryAddress) == 0 || len(registryUsername) == 0 || len(registryPassword) == 0 {
			err = fmt.Errorf("missing private registry data: ")
			if len(registryAddress) == 0 {
				err = errors.Wrapf(err, "missing %s", keyDockerprivateRegistryAddress)
			}
			if len(registryUsername) == 0 {
				err = errors.Wrapf(err, "missing %s", keyDockerprivateRegistryUsername)
			}
			if len(registryPassword) == 0 {
				err = errors.Wrapf(err, "missing %s", keyDockerprivateRegistryPassword)
			}
			return nil, err
		}
	}
	// Creating 
	registryTplData := map[string]interface{}{
		"RegistryAddress": registryAddress,
		"RegistryUsername": registryUsername,
		"RegistryPassword": registryPassword,
	} 
	registryValues, err := ExecuteTemplateToValues(RegistryValuesTpl, registryTplData)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error to parse docker registry values"))
  }

	return registryValues, nil
}

func (o *CfApply) ApplyCodefresh() error {

	registryValues, err := o.GetDockerRegistryVars()
	//_, err := o.AddDockerRegistryVars(vals)
	if err != nil {
		return errors.Wrapf(err, "Failed to parse docker registry values")
	}
	o.vals = MergeMaps(o.vals, registryValues)
	valsX := objx.New(o.vals)
	
	// If a release does not exist add seeded jobs
	histClient := helm.NewHistory(o.cfg)
	histClient.Max = 1
	if _, err := histClient.Run(codefreshHelmReleaseName); err == driver.ErrReleaseNotFound {
		seedJobsValues := map[string]interface{}{
			"global": map[string]interface{}{
				"seedJobs": true,
				"certsJobs": true,
			},
		}
		o.vals = MergeMaps(o.vals, seedJobsValues)
	}

	valuesTplResult, err := ExecuteTemplate(ValuesTpl, o.vals)
	if err != nil {
		return errors.Wrapf(err, "Failed to generate values.yaml")
	}

	valuesYamlPath := path.Join(GetAssetsDir(o.ConfigFile), "values.yaml")
	err = ioutil.WriteFile(valuesYamlPath, []byte(valuesTplResult), 0644)
	if err != nil {
		return errors.Wrapf(err, "Failed to write %s ", valuesYamlPath)
	}
	fmt.Printf("values.yaml has been generated in %s\n", valuesYamlPath)

	cfResourceTplResult, err := ExecuteTemplate(CfResourceTpl, o.vals)
	if err != nil {
		return errors.Wrapf(err, "Failed to generate codefresh-resource.yaml")
	}
	cfResourceYamlPath := path.Join(GetAssetsDir(o.ConfigFile), "codefresh-resource.yaml")
	err = ioutil.WriteFile(cfResourceYamlPath, []byte(cfResourceTplResult), 0644)
	if err != nil {
		return errors.Wrapf(err, "Failed to write %s ", cfResourceYamlPath)
	}
	fmt.Printf("codefresh-resource.yaml is generated in %s\n", cfResourceYamlPath)

    installerType := valsX.Get(keyInstallerType).String()
    if installerType == installerTypeOperator {
		
		// Deploy Codefresh Operator with wait first
		operatorChartValues := valsX.Get(keyOperatorChartValues).MSI(map[string]interface{}{})
		helmWaitBak := o.Helm.Wait
		o.Helm.Wait = true
		_, err = DeployHelmRelease(
					operatorHelmReleaseName, 
					operatorHelmChartName, 
					operatorChartValues, 
					o.cfg, 
					o.Helm,
				)
		if err != nil {
			return errors.Wrapf(err, "Failed to deploy operator chart")	
		}
		o.Helm.Wait = helmWaitBak

		// Update Codefresh resourtc 
        cfResourceYamlReader, err := os.Open(cfResourceYamlPath)
        if err != nil {
            return errors.Wrapf(err, "Failed to read %s ", cfResourceYamlPath)
        }
        cfResources, err := o.cfg.KubeClient.Build(cfResourceYamlReader, true)
        if err != nil {
            return errors.Wrapf(err, "Failed to write %s ", cfResourceYamlPath)
        }
        fmt.Printf("applying %s\n %v", cfResourceYamlPath, cfResources)
        err = cfResources.Visit(func(info *resource.Info, err error) error {
            if err != nil {
                return err
            }

            helper := resource.NewHelper(info.Client, info.Mapping)
            _, err = helper.Replace(info.Namespace, info.Name, true, info.Object)
            return err
		})
		if err != nil {
			return errors.Wrapf(err, "Failed to apply %s\n", cfResourceYamlPath)
		}   
    } else if installerType == installerTypeHelm {
		// first we will error if operator chart is installed:
		histClient := helm.NewHistory(o.cfg)
		histClient.Max = 1
		if operatorRelease, _ := histClient.Run(operatorHelmReleaseName); operatorRelease != nil {
			return fmt.Errorf("Error: Codefresh operator release is running. It is incomplatible with helm install type")
		}
		codefreshHelChartName := valsX.Get(keyCodefreshHelmChart).Str("codefresh.tgz")
		_, err = DeployHelmRelease(
			codefreshHelmReleaseName, 
			codefreshHelChartName, 
			o.vals, 
			o.cfg, 
			o.Helm,
		)
		if err != nil {
			return errors.Wrapf(err, "Failed to deploy operator chart")	
		}		
	} else { 
		return fmt.Errorf("Error: unknown instraller type %s", installerType)
	}

	fmt.Printf("\nCodefresh has been deployed to namespace %s\n", o.Helm.Namespace)
    return nil
}

// ValuesTpl is a template to format final helm values
var ValuesTpl = `
{{ . | toYaml }}

`

//CfResourceTpl template to final Codefresh custom resource
var CfResourceTpl = `
apiVersion: codefresh.io/v1alpha1
kind: Codefresh
metadata:
  name: cf
  namespace: {{ .kubernetes.namespace }}
spec:
{{ . | toYaml | indent 2 }}
`

// RegistryValuesTpl template
var RegistryValuesTpl = `
{{ $auth := ((printf "%s:%s" .RegistryUsername .RegistryPassword ) | b64enc) }}
dockerconfigjson:
  auths:
    {{.RegistryAddress | toString }}:
      auth: {{ $auth }}
global:
  {{- if .docker.usePrivateRegistry }}
  privateRegistry: true
  dockerRegistry: {{ printf "%s/" .RegistryAddress }}
  {{- end }}
  dockerconfigjson:
    auths:
      {{.RegistryAddress | toString }}:
        auth: {{ $auth }}
cfui:
  dockerconfigjson:
    auths:
      {{.RegistryAddress | toString }}:
        auth: {{ $auth }}
runtime-environment-manager:
  dockerconfigjson:
    auths:
      {{.RegistryAddress | toString }}:
        auth: {{ $auth }}
onboarding-status:
  dockerconfigjson:
    auths:
      {{.RegistryAddress | toString }}:
        auth: {{ $auth }}
cfanalytic:
  dockerconfigjson:
    auths:
      {{.RegistryAddress | toString }}:
        auth: {{ $auth }}
`