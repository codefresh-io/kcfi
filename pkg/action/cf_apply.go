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
	"time"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/stretchr/objx"

	helm "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/postrender"
)

var (
	keyRelease = "installer.release"
	
	keyDockerCodefreshRegistrySa = "docker.codefreshRegistrySa"
	keyDockerUsePrivateRegistry = "docker.usePrivateRegistry"
	keyDockerprivateRegistryAddress = "docker.privateRegistry.address"
	keyDockerprivateRegistryUsername = "docker.privateRegistry.username"
	keyDockerprivateRegistryPassword = "docker.privateRegistry.password"
)

// CfApply is an action to creat or update Codefresh
type CfApply struct {
	ConfigFile string
	cfg *helm.Configuration

	// Helm Install/Upgrader optional parameters
	helm.ChartPathOptions

	// Install is a purely informative flag that indicates whether this upgrade was done in "install" mode.
	//
	// Applications may use this to determine whether this Upgrade operation was done as part of a
	// pure upgrade (Upgrade.Install == false) or as part of an install-or-upgrade operation
	// (Upgrade.Install == true).
	//
	// Setting this to `true` will NOT cause `Upgrade` to perform an install if the release does not exist.
	// That process must be handled by creating an Install action directly. See cmd/upgrade.go for an
	// example of how this flag is used.
	Install bool
	// Devel indicates that the operation is done in devel mode.
	Devel bool
	// Namespace is the namespace in which this operation should be performed.
	Namespace string
	// SkipCRDs skips installing CRDs when install flag is enabled during upgrade
	SkipCRDs bool
	// Timeout is the timeout for this operation
	Timeout time.Duration
	// Wait determines whether the wait operation should be performed after the upgrade is requested.
	Wait bool
	// DisableHooks disables hook processing if set to true.
	DisableHooks bool
	// DryRun controls whether the operation is prepared, but not executed.
	// If `true`, the upgrade is prepared but not performed.
	DryRun bool
	// Force will, if set to `true`, ignore certain warnings and perform the upgrade anyway.
	//
	// This should be used with caution.
	Force bool
	// ResetValues will reset the values to the chart's built-ins rather than merging with existing.
	ResetValues bool
	// ReuseValues will re-use the user's last supplied values.
	ReuseValues bool
	// Recreate will (if true) recreate pods after a rollback.
	Recreate bool
	// MaxHistory limits the maximum number of revisions saved per release
	MaxHistory int
	// Atomic, if true, will roll back on failure.
	Atomic bool
	// CleanupOnFail will, if true, cause the upgrade to delete newly-created resources on a failed update.
	CleanupOnFail bool
	// SubNotes determines whether sub-notes are rendered in the chart.
	SubNotes bool
	// Description is the description of this operation
	Description string
	// PostRender is an optional post-renderer
	//
	// If this is non-nil, then after templates are rendered, they will be sent to the
	// post renderer before sending to the Kuberntes API server.
	PostRenderer postrender.PostRenderer
	// DisableOpenAPIValidation controls whether OpenAPI validation is enforced.
	DisableOpenAPIValidation bool
}



// NewCfApply creates object
func NewCfApply(cfg *helm.Configuration) *CfApply {
	return &CfApply{
		cfg: cfg,
	}
}

// AddDockerRegistryVars - adds docker registry to vals
func (o *CfApply) AddDockerRegistryVars (vals map[string]interface{}) error {
	
	var registryAddress, registryUsername, registryPassword string
	var err error
	valsX := objx.New(vals)
	usePrivateRegistry := valsX.Get(keyDockerUsePrivateRegistry).Bool(false)
	if !usePrivateRegistry {
		// using Codefresh Enterprise registry
		registryAddress = "gcr.io"
		registryUsername = "_json_key"
		cfRegistrySaVal := valsX.Get(keyDockerCodefreshRegistrySa).Str("sa.json")
		cfRegistrySaPath := path.Join(filepath.Dir(o.ConfigFile), cfRegistrySaVal)
    registryPasswordB, err := ioutil.ReadFile(cfRegistrySaPath)
    if err != nil {
        return errors.Wrap(err, fmt.Sprintf("cannot read %s", cfRegistrySaPath))
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
			return err
		}
	}
	// Creating 
	registryTplData := map[string]interface{}{
		"RegistryAddress": registryAddress,
		"DockerCfg": fmt.Sprintf("%s:%s", registryUsername, registryPassword),
	} 
	registryValues, err := ExecuteTemplateToValues(RegistryValuesTpl, registryTplData)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error to parse docker registry values"))
  }
  vals = MergeMaps(vals, registryValues)
	return nil
}

// Run the action
func (o *CfApply) Run(vals map[string]interface{}) error {
	fmt.Printf("Applying Codefresh configuration from %s\n", o.ConfigFile)
	
	err := o.AddDockerRegistryVars(vals)
	if err != nil {
		return errors.Wrapf(err, "Failed to parse docker registry values")
	}
	
	return nil
}