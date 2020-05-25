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
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/stretchr/objx"

	helm "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/storage/driver"
	kerrors "k8s.io/apimachinery/pkg/api/errors"

	c "github.com/codefresh-io/kcfi/pkg/config"
	"k8s.io/cli-runtime/pkg/resource"
)

// GetDockerRegistryVars - calculater docker registry vals
func (o *CfApply) GetDockerRegistryVars() (map[string]interface{}, error) {

	var registryAddress, registryUsername, registryPassword string
	var err error
	valsX := objx.New(o.vals)
	usePrivateRegistry := valsX.Get(c.KeyImagesUsePrivateRegistry).Bool(false)
	if !usePrivateRegistry {
		// using Codefresh Enterprise registry
		registryAddress = c.CfRegistryAddress
		registryUsername = c.CfRegistryUsername
		cfRegistrySaVal := valsX.Get(c.KeyImagesCodefreshRegistrySa).Str("sa.json")
		cfRegistrySaPath := o.filePath(cfRegistrySaVal)
		registryPasswordB, err := ioutil.ReadFile(cfRegistrySaPath)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("cannot read %s", cfRegistrySaPath))
		}
		registryPassword = string(registryPasswordB)
	} else {
		registryAddress = valsX.Get(c.KeyImagesPrivateRegistryAddress).String()
		registryUsername = valsX.Get(c.KeyImagesPrivateRegistryUsername).String()
		registryPassword = valsX.Get(c.KeyImagesPrivateRegistryPassword).String()
		if len(registryAddress) == 0 || len(registryUsername) == 0 || len(registryPassword) == 0 {
			err = fmt.Errorf("missing private registry data: ")
			if len(registryAddress) == 0 {
				err = errors.Wrapf(err, "missing %s", c.KeyImagesPrivateRegistryAddress)
			}
			if len(registryUsername) == 0 {
				err = errors.Wrapf(err, "missing %s", c.KeyImagesPrivateRegistryUsername)
			}
			if len(registryPassword) == 0 {
				err = errors.Wrapf(err, "missing %s", c.KeyImagesPrivateRegistryPassword)
			}
			return nil, err
		}
	}
	// Creating
	registryTplData := map[string]interface{}{
		"RegistryAddress":  registryAddress,
		"RegistryUsername": registryUsername,
		"RegistryPassword": registryPassword,
	}
	registryValues, err := ExecuteTemplateToValues(RegistryValuesTpl, registryTplData)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error to parse docker registry values"))
	}

	return registryValues, nil
}

func (o *CfApply) applyDbInfra() error {
	valsX := objx.New(o.vals)
	if ! valsX.Get(c.KeyDbInfraEnabled).Bool(false) {
		debug("%s is not enabled", c.KeyDbInfraEnabled)
		return nil
	}
	info("%s is enabled", c.KeyDbInfraEnabled)

	dbInfraConfig, err := ReadYamlFile(o.filePath(c.DbInfraConfigFile))
	if err != nil {
		return errors.Wrapf(err, "failed to parse db-infra config file %s", c.DbInfraConfigFile)
	}
	
	dbInfraConfig = MergeMaps(dbInfraConfig, valsX.Get(c.KeyDbInfra).MSI(map[string]interface{}{}))
	dbInfraConfigX := objx.New(dbInfraConfig)
	dbInfraReleaseName := dbInfraConfigX.Get(c.KeyHelmRelease).String()

	// Checking if dbInfra is already installed
	dbInfraInstalled := IsHelmReleaseInstalled(dbInfraReleaseName, o.cfg)
	codefreshInstalled := IsHelmReleaseInstalled(c.CodefreshReleaseName, o.cfg)

	if codefreshInstalled && !dbInfraInstalled {
		return fmt.Errorf("db-infra release %s is not installed", dbInfraReleaseName)
	}

	// Merging values/db-infra into main values
	mainConfigChange, err := ReadYamlFile(o.filePath(c.DbInfraMainConfigChangeValuesFile))
	if err != nil {
		return errors.Wrapf(err, "failed to parse db-infra values file %s", c.DbInfraMainConfigChangeValuesFile)
	}
	debug("merging db-infra config change file %s", c.DbInfraMainConfigChangeValuesFile)
	o.vals =  MergeMaps(o.vals, mainConfigChange)

	dbInfraUpgrade := valsX.Get(c.KeyDbInfraUpgrade).Bool(false)
	debug("dbInfraUpgrade = %t", dbInfraUpgrade)
	if ( !codefreshInstalled && !dbInfraInstalled ) || dbInfraUpgrade {
		info("Installing db-infra release")
		helmAtomicSave := o.Helm.Atomic 
		helmWaitSave := o.Helm.Wait
		o.Helm.Atomic = true
		o.Helm.Wait = true
		defer func(){
			o.Helm.Atomic = helmAtomicSave
			o.Helm.Wait = helmWaitSave
		}()
		dbInfraRelease, err := DeployHelmRelease(
			dbInfraReleaseName,
			c.DbInfraHelmChartName,
			dbInfraConfig,
			o.cfg,
			o.Helm,
		)
		if err != nil {
			return errors.Wrapf(err, "Failed to deploy db-infra chart")
		}
		PrintHelmReleaseInfo(dbInfraRelease, c.Debug)

	}

	return nil
}

// ValidateCodefresh validates Codefresh config values
func (o *CfApply) ValidateCodefresh() error {
	valsX := objx.New(o.vals)

	// 1. Validate appUrl
	appUrl := valsX.Get(c.KeyAppUrl).String()
	if appUrl == "" {
		return fmt.Errorf("Missing %s", c.KeyAppUrl)
	}
	return nil
}

func (o *CfApply) ApplyCodefresh() error {

	// Calculating addional configurations
	valsX := objx.New(o.vals)

	//--- Docker Registry secret
	registryValues, err := o.GetDockerRegistryVars()
	if err != nil {
		return errors.Wrapf(err, "Failed to parse docker registry values")
	}
	usePrivateRegistry := valsX.Get(c.KeyImagesUsePrivateRegistry).Bool(false)
	if usePrivateRegistry {
		privateRegistryAddress := valsX.Get(c.KeyImagesPrivateRegistryAddress).String()
		privateRegistryGlobalValues := map[string]interface{}{
			"global": map[string]interface{}{
				"privateRegistry":  true,
				"dockerRegistry": privateRegistryAddress + "/",
			},
		}
		o.vals = MergeMaps(o.vals, privateRegistryGlobalValues)
	}
	o.vals = MergeMaps(o.vals, registryValues)

	//--- WebTls Values
	if ! valsX.Get(c.KeyTlsSelfSigned).Bool(true) {
		webTlsValues, err := ExecuteTemplateToValues(WebTlsValuesTpl, o.vals)
		if err != nil {
			return errors.Wrapf(err, "Failed to generate values.yaml")
		}
		o.vals = MergeMaps(o.vals, webTlsValues)
	}

	// Db Infra
	err = o.applyDbInfra()
	if err != nil {
		return errors.Wrapf(err, "Failed apply db-infra")
	}

	//--- If a release does not exist add seeded jobs
	histClient := helm.NewHistory(o.cfg)
	histClient.Max = 1
	if _, err := histClient.Run(codefreshHelmReleaseName); err == driver.ErrReleaseNotFound {
		seedJobsValues := map[string]interface{}{
			"global": map[string]interface{}{
				"seedJobs":  true,
				"certsJobs": true,
			},
		}
		o.vals = MergeMaps(o.vals, seedJobsValues)
	}

	// Validating Configurations
	if err = o.ValidateCodefresh(); err != nil {
		return errors.Wrapf(err, "Configuration is not valid")
	}

	// Rendering values.yaml and codefresh-resource.yaml
	valuesTplResult, err := ExecuteTemplate(ValuesTpl, o.vals)
	if err != nil {
		return errors.Wrapf(err, "Failed to generate values.yaml")
	}

	valuesYamlPath := path.Join(GetAssetsDir(o.ConfigFile), "values.yaml")
	err = ioutil.WriteFile(valuesYamlPath, []byte(valuesTplResult), 0644)
	if err != nil {
		return errors.Wrapf(err, "Failed to write %s ", valuesYamlPath)
	}
	info("values.yaml has been generated in %s\n", valuesYamlPath)

	cfResourceTplResult, err := ExecuteTemplate(CfResourceTpl, o.vals)
	if err != nil {
		return errors.Wrapf(err, "Failed to generate codefresh-resource.yaml")
	}
	cfResourceYamlPath := path.Join(GetAssetsDir(o.ConfigFile), "codefresh-resource.yaml")
	err = ioutil.WriteFile(cfResourceYamlPath, []byte(cfResourceTplResult), 0644)
	if err != nil {
		return errors.Wrapf(err, "Failed to write %s ", cfResourceYamlPath)
	}
	info("codefresh-resource.yaml is generated in %s\n", cfResourceYamlPath)

	//--- Deploying
    installerType := valsX.Get(c.KeyInstallerType).String()
    if installerType == installerTypeOperator {
		// Deploy Codefresh Operator with wait first
		operatorChartValues := valsX.Get(c.KeyOperatorChartValues).MSI(map[string]interface{}{})
		operatorChartValues = MergeMaps(operatorChartValues, registryValues)
		if usePrivateRegistry {
			operatorChartValues = MergeMaps(operatorChartValues, map[string]interface{}{
				c.KeyDockerRegistry: valsX.Get(c.KeyImagesPrivateRegistryAddress).String(),
			})
		}
		if valsX.Get(c.KeyOperatorSkipCRD).Bool(false) {
			o.Helm.SkipCRDs = true
		}
		o.Helm.Atomic = true
		o.Helm.Wait = true
		operatorRelease, err := DeployHelmRelease(
			operatorHelmReleaseName,
			operatorHelmChartName,
			operatorChartValues,
			o.cfg,
			o.Helm,
		)
		if err != nil {
			return errors.Wrapf(err, "Failed to deploy operator chart")
		}
		PrintHelmReleaseInfo(operatorRelease, c.Debug)

		// Update Codefresh resource 
        cfResourceYamlReader, err := os.Open(cfResourceYamlPath)
        if err != nil {
            return errors.Wrapf(err, "Failed to read %s ", cfResourceYamlPath)
        }
        cfResources, err := o.cfg.KubeClient.Build(cfResourceYamlReader, true)
        if err != nil {
            return errors.Wrapf(err, "Failed to write %s ", cfResourceYamlPath)
        }
		info("applying %s\n %v", cfResourceYamlPath, cfResources)
		if o.Helm.DryRun {
			info("\n\nDryRun Mode - Codefresh Resource Definition is generatest in %s", cfResourceYamlPath)
			return nil
		}
		err = cfResources.Visit(func(info *resource.Info, err error) error {
			if err != nil {
				return err
			}
			helper := resource.NewHelper(info.Client, info.Mapping)
			if _, err = helper.Get(info.Namespace, info.Name, info.Export); err != nil {
				if !kerrors.IsNotFound(err) {
					return errors.Wrapf(err, fmt.Sprintf("retrieving current configuration of:\n%s\nfrom server for:", info.String()))
				}
				_, err = helper.Create(info.Namespace, true, info.Object)
			} else {
				_, err = helper.Replace(info.Namespace, info.Name, true, info.Object)
			}
			return err
		})
		if err != nil {
			return errors.Wrapf(err, "Failed to apply %s\n", cfResourceYamlPath)
		}
	} else if installerType == installerTypeHelm {
		// first we will error if operator chart is installed:
		if operatorReleaseInstalled := IsHelmReleaseInstalled(operatorHelmReleaseName, o.cfg); operatorReleaseInstalled {
			return fmt.Errorf("Error: Codefresh operator release is running. It is incomplatible with helm install type")
		}
		codefreshHelChartName := valsX.Get(c.KeyHelmChart).Str("codefresh")
		codefreshRelease, err := DeployHelmRelease(
			codefreshHelmReleaseName,
			codefreshHelChartName,
			o.vals,
			o.cfg,
			o.Helm,
		)
		if err != nil {
			return errors.Wrapf(err, "Failed to deploy operator chart")
		}
		PrintHelmReleaseInfo(codefreshRelease, false)
	} else {
		return fmt.Errorf("Error: unknown instraller type %s", installerType)
	}

	info("\nCodefresh has been deployed to namespace %s\n", o.Helm.Namespace)
	return nil
}

func (o *CfApply) ApplyBackupMgr() error {
	valsX := objx.New(o.vals)

	mongoURI := valsX.Get(c.KeyBkpManagerMongoURI).Str()
	if mongoURI == "" {
		info("Mongo URI is not specified, trying to get it automatically from the installed Codefresh release")
		var err error
		mongoURI, err = o.getMongoURIFromRelease()
		if err != nil {
			return err
		}
	}

	debug("Mongo URI is: %s", mongoURI)
	valsX.Set(c.KeyBkpManagerMongoURI, mongoURI)

	installerType := valsX.Get(c.KeyInstallerType).String()
	if installerType == installerTypeHelm {
		helmChartName := valsX.Get(c.KeyHelmChart).String()
		helmReleaseName := valsX.Get(c.KeyHelmRelease).Str(kindBackupManager)
		rel, err := DeployHelmRelease(
			helmReleaseName,
			helmChartName,
			valsX,
			o.cfg,
			o.Helm,
		)
		if err != nil {
			return errors.Wrapf(err, "Failed to deploy %s chart", helmChartName)
		}
		PrintHelmReleaseInfo(rel, c.Debug)
		info("\n%s has been deployed to namespace %s\n", helmReleaseName, o.Helm.Namespace)
		return nil
	}
	return fmt.Errorf("Wrong installer type %s", installerType)
}

func (o *CfApply) getMongoURIFromRelease() (string, error) {

	cfRelVals, err := GetReleaseValues(codefreshHelmReleaseName, o.cfg)
	if err != nil {
		return "", err
	}

	cfRelValsX := objx.New(cfRelVals)

	runtimeMongoURI := cfRelValsX.Get(c.KeyGlobalMongoURI).Str()
	mongoRootUser := cfRelValsX.Get(c.KeyGlobalMongoRootUser).Str()
	mongoRootPassword := cfRelValsX.Get(c.KeyGlobalMongoRootPassword).Str()

	if !strings.Contains(runtimeMongoURI, "@") || mongoRootUser == "" || mongoRootPassword == "" {
		return "", fmt.Errorf("Failed to get the mongo URI value from an existing release")
	}

	rootMongoURI := "mongodb://" + mongoRootUser + ":" + mongoRootPassword + "@" + strings.Split(runtimeMongoURI, "@")[1]
	return rootMongoURI, nil
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

// WebTlsValuesTpl template
var WebTlsValuesTpl = `
ingress:
  webTlsSecretName: "star.codefresh.io"
nomios:
  ingress:
    webTlsSecretName: "star.codefresh.io"
webTLS:
  secretName: star.codefresh.io
  key: |
{{ getFileWithBaseDir .tls.key .BaseDir | indent 4}}
  cert: |
{{ getFileWithBaseDir .tls.cert .BaseDir | indent 4}}

cfui:
  webTLS:
    key: |
{{ getFileWithBaseDir .tls.key .BaseDir | indent 6}}
    cert: |
{{ getFileWithBaseDir .tls.cert .BaseDir | indent 6}}
`
