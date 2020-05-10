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
	"os"
	"fmt"
	"strings"
	"regexp"
	"io/ioutil"
	"path"
	"path/filepath"
	"github.com/codefresh-io/kcfi/pkg/embeded/stage"
)

const (
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
	keyOperatorSkipCRD = "metadata.installer.operator.skipCRD"

	operatorHelmReleaseName = "cf-onprem-operator"
	operatorHelmChartName = "codefresh-operator"
	
	keyCodefreshHelmChart = "metadata.installer.helm.chart"
	codefreshHelmReleaseName = "cf"

	keyNamespace = "kubernetes.namespace"

	keyBaseDir = "BaseDir"
	keyTlsSelfSigned = "tls.selfSigned"
	keyTlsCert = "tls.cert"
	keyTlsKey = "tls.key"

	keyAppUrl = "global.appUrl"
)

// CfInit is an action to create Codefresh config stage directory
type CfInit struct {
	ProductName string
	StageDir string
}

// NewCfInit creates object
func NewCfInit(productName, stageDir string) *CfInit {
	return &CfInit{
		ProductName: productName,
		StageDir: stageDir,
	}
}

// Run the action
func (o *CfInit) Run() error {
	var isValidProduct bool
	for _, name := range StageDirsList() {
		if o.ProductName == name {
			isValidProduct = true
			break
		}
	}
	if !isValidProduct {
	   return fmt.Errorf("Unknown product %s", o.ProductName)
	}

	var err error
	var restoreDir string
	if len(o.StageDir) == 0 {
		o.StageDir, err = os.Getwd()
		if err != nil {
			return err
		}
		restoreDir = path.Join(o.StageDir, o.ProductName)
	} else {
		restoreDir = o.StageDir
	}
	
	fmt.Printf("Creating stage directory %s\n", restoreDir )
	if dirList, err := ioutil.ReadDir(restoreDir); err == nil && len(dirList) > 0 {
		return fmt.Errorf("Directory %s is already exists and not empty", o.ProductName)
	}
	return restoreStageAssets(restoreDir, o.ProductName)
}

// restoreStageAssets restores an asset with replacing first folder under the given directory recursively
func restoreStageAssets(dir, name string) error {
	children, err := stage.AssetDir(name)
	// File
	if err != nil {
		return restoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = restoreStageAssets(dir, fmt.Sprintf("%s/%s", name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

// restoreAsset restores an asset under the given directory with removing first folder
func restoreAsset(dir, name string) error {
	data, err := stage.Asset(name)
	if err != nil {
		return err
	}
	info, err := stage.AssetInfo(name)
	if err != nil {
		return err
	}

	stageFileNameReplaceRe, _ := regexp.Compile(`^(.*?/)(.*$)$`)
	stageFileName := stageFileNameReplaceRe.ReplaceAllString(name, "$2")
	err = os.MkdirAll(_filePath(dir, filepath.Dir(stageFileName)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, stageFileName), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, stageFileName), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

// StageDirsList - returns list of registered staging dir
func StageDirsList() []string {	
	var stageDirsList []string
	var stageName string
	stageDirsMap := make(map[string]int)
	stageNameReplaceRe, _ := regexp.Compile(`^(.*?)/(.*$)$`)

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