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
	"github.com/pkg/errors"
	"github.com/stretchr/objx"
	helm "helm.sh/helm/v3/pkg/action"
	c "github.com/codefresh-io/kcfi/pkg/config"
)

// CfApply is an action to create or update Codefresh
type CfApply struct {
	ConfigFile string
	vals map[string]interface{}
	cfg *helm.Configuration
	Helm *helm.Upgrade
}

// NewCfApply creates object
func NewCfApply(cfg *helm.Configuration) *CfApply {
	return &CfApply{
		cfg:  cfg,
		Helm: helm.NewUpgrade(cfg),
	}
}

// Run the action
func (o *CfApply) Run(vals map[string]interface{}) error {
	info("Applying Codefresh configuration from %s\n", o.ConfigFile)
	// info("Applying Codefresh configuration from %s\n", o.ConfigFile)
	o.vals = vals
	valsX := objx.New(vals)
	
	// Includes additional value files
	if valsX.Has(c.KeyInclude) {
		var includeFiles []string
		// cfgX.Get(c.KeyInclude).StrSlice() - not working, returns empty
		includeFilesI := valsX.Get(c.KeyInclude).Data()
		fileNamesI, includeIsList := includeFilesI.([]interface{}) 
		if !includeIsList {
			return fmt.Errorf("Error: %s - %v is not a list", c.KeyInclude, includeFilesI)
		}
		for _, f := range fileNamesI {
			if str, isStr := f.(string); isStr {
				includeFiles = append(includeFiles, str)
			} else {
				return fmt.Errorf("Error: %s - %v is not a string", c.KeyInclude, f)
			}
		}
		debug("Processed include: %v - %v", includeFilesI, includeFiles)

		for _, fileName := range includeFiles {
			includeConfig, err := ReadYamlFile(path.Join(filepath.Dir(o.ConfigFile), fileName))
			if err != nil {
				return errors.Wrapf(err, "failed to parse included values file %s", fileName)
			}
			debug("merging included config change file %s", fileName)
			o.vals =  MergeMaps(o.vals, includeConfig)
		}
	}

	kind := valsX.Get(c.KeyKind).String(); 
	baseDir := filepath.Dir(o.ConfigFile)
	o.vals[c.KeyBaseDir] = baseDir

	switch kind {
	case kindCodefresh:
		return o.ApplyCodefresh()
	case "":
		return fmt.Errorf("Please specifiy the installer kind")
	default:
		installerType := valsX.Get(c.KeyInstallerType).String()
		if installerType == installerTypeHelm {
			helmChartName := valsX.Get(c.KeyHelmChart).String()
			helmReleaseName := valsX.Get(c.KeyHelmRelease).Str(kind)
			rel, err := DeployHelmRelease(
				helmReleaseName,
				helmChartName,
				o.vals,
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
}

func (o *CfApply) filePath(fileName string) string {
	return path.Join(filepath.Dir(o.ConfigFile), fileName)
}
