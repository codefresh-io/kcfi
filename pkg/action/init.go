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
	"regexp"
	"path"
	"path/filepath"
	"github.com/codefresh-io/kcfi/pkg/embeded/stage"
)

const (
	// DefaultConfigFileName - 
	DefaultConfigFileName = "config.yaml"

	// AssetsDir - folder name where we save kubernetes and helm assets
	AssetsDir = "assets"

	CodefreshReleaseName = "cf"
	OperatorReleaseName = "cf-onprem-operator"

	keyKind = "metadata.kind"
	kindCodefresh = "codefresh"
	kindK8sAgent = "k8sAgent"
	kindVenona = "venona"
	
	keyDockerCodefreshRegistrySa = "docker.codefreshRegistrySa"
	keyDockerUsePrivateRegistry = "docker.usePrivateRegistry"
	keyDockerprivateRegistryAddress = "docker.privateRegistry.address"
	keyDockerprivateRegistryUsername = "docker.privateRegistry.username"
	keyDockerprivateRegistryPassword = "docker.privateRegistry.password"

	keyRelease = "metadata.installer.release"	
	keyInstallerType = "metadata.installer.type"
	installerTypeOperator = "operator"
	installerTypeHelm = "helm"

	keyOperatorChartValues = "metadata.installer.operator"

	operatorHelmReleaseName = "cf-onprem-operator"
	operatorHelmChartName = "codefresh-operator"
	
	keyCodefreshHelmChart = "metadata.installer.helm.chart"
	codefreshHelmReleaseName = "cf"

	keyNamespace = "kubernetes.namespace"

)

// CfInit is an action to create Codefresh config stage directory
type CfInit struct {
	productName string
	stageDir string
}

// NewCfInit creates object
func NewCfInit(productName, stageDir string) *CfInit {
	return &CfInit{
		productName: productName,
		stageDir: stageDir,
	}
}

// Run the action
func (o *CfInit) Run() error {
	var isValidProduct bool
	for _, name := range StageDirsList() {
		if o.productName == name {
			isValidProduct = true
			break
		}
	}
	if !isValidProduct {
	   return fmt.Errorf("Unknown product %s", o.productName)
	}
	fmt.Printf("Creating stage directory in %s\n", path.Join(o.stageDir, o.productName))
	return stage.RestoreAssets(o.stageDir, o.productName)
}

// StageDirsList - returns list of registered staging dir
func StageDirsList() []string {	
	var stageDirsList []string
	var stageName string
	stageDirsMap := make(map[string]int)
	stageNameReplaceRe, _ := regexp.Compile(`^(.*)/(.*$)$`)

	for _, name := range stage.AssetNames() {
		stageName = stageNameReplaceRe.ReplaceAllString(name, "$1")
		if _, stageNameListed := stageDirsMap[stageName]; !stageNameListed {
			stageDirsMap[stageName] = 1
			stageDirsList = append(stageDirsList, stageName)
		}
	}
	return stageDirsList
}

// GetAssetsDir - retur assets dir
func GetAssetsDir(configFile string) string {
	return path.Join(filepath.Dir(configFile), AssetsDir)
}

func debug(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}