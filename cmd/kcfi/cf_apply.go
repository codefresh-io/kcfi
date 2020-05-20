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

package main

import (
	"fmt"
	"io"
	"time"
	"os"
	"filepath"
	"github.com/spf13/cobra"

	"github.com/codefresh-io/kcfi/pkg/action"
	"helm.sh/helm/v3/cmd/helm/require"
	helm "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli/output"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"

	c "github.com/codefresh-io/kcfi/pkg/config"
)

const cfApplyDesc = `
This command deploys Codefresh product with parameters defined in configuration file
   kcfi deploy [-c|--config /path/to/codefresh/config.yaml ]
by default it looks for config.yaml in current directory,
`

func cfApplyCmd(cfg *helm.Configuration, out io.Writer) *cobra.Command {
	client := action.NewCfApply(cfg)
	valueOpts := &values.Options{}
	var outfmt output.Format
	//var createNamespace bool

	cmd := &cobra.Command{
		Use:     "deploy",
		Short:   "install/upgrade/reconfigure Codefresh",
		Aliases: []string{"apply", "install", "upgrade"},
		Long:    cfApplyDesc,
		Args:    require.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// merging configFile with valueOpts
			valueOpts.ValueFiles = append([]string{client.ConfigFile}, valueOpts.ValueFiles...)
			var baseDir string
			if fInfo, err := os.Stat(client.ConfigFile); err == nil && !fInfo.IsDir(){
				valueOpts.ValueFiles = []string{client.ConfigFile}
				baseDir = filepath.Dir(client.ConfigFile)
			} else {
				return fmt.Errorf("%s is not a valid file", client.ConfigFile)
			}
			valueOpts.Values = append(valueOpts.Values, fmt.Sprintf("%s=%s", c.KeyBaseDir, baseDir))
			
			valueOpts.Values = append(valueOpts.Values, fmt.Sprintf("%s=%s", c.KeyKubeNamespace, configuredNamespace))
			vals, err := valueOpts.MergeValues(getter.All(settings))
			if err != nil {
				return err
			}
			client.Helm.Namespace = configuredNamespace

			return client.Run(vals)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&client.ConfigFile, flagConfig, "c", defaultConfigFileName(), "Codefresh config file")
	//f.BoolVar(&createNamespace, "create-namespace", false, "if --install is set, create the release namespace if not present")
	//f.BoolVarP(&client.Install, "install", "i", true, "if a release by this name doesn't already exist, run an install")
	//f.BoolVar(&client.Helm.Devel, "devel", false, "use development versions, too. Equivalent to version '>0.0.0-0'. If --version is set, this is ignored")
	f.BoolVar(&client.Helm.DryRun, "dry-run", false, "simulate an upgrade")
	//f.BoolVar(&client.Helm.Recreate, "recreate-pods", false, "performs pods restart for the resource if applicable")
	//f.MarkDeprecated("recreate-pods", "functionality will no longer be updated. Consult the documentation for other methods to recreate pods")
	//f.BoolVar(&client.Helm.Force, "force", false, "force resource updates through a replacement strategy")
	f.BoolVar(&client.Helm.DisableHooks, "no-hooks", false, "disable pre/post upgrade hooks")
	//f.BoolVar(&client.Helm.DisableOpenAPIValidation, "disable-openapi-validation", false, "if set, the upgrade process will not validate rendered templates against the Kubernetes OpenAPI Schema")
	//f.BoolVar(&client.Helm.SkipCRDs, "skip-crds", false, "if set, no CRDs will be installed when an upgrade is performed with install flag enabled. By default, CRDs are installed if not already present, when an upgrade is performed with install flag enabled")
	f.DurationVar(&client.Helm.Timeout, "timeout", 300*time.Second, "time to wait for any individual Kubernetes operation (like Jobs for hooks)")
	f.BoolVar(&client.Helm.ResetValues, "reset-values", false, "when upgrading, reset the values to the ones built into the chart")
	f.BoolVar(&client.Helm.ReuseValues, "reuse-values", false, "when upgrading, reuse the last release's values and merge in any overrides from the command line via --set and -f. If '--reset-values' is specified, this is ignored")
	f.BoolVar(&client.Helm.Wait, "wait", false, "if set, will wait until all Pods, PVCs, Services, and minimum number of Pods of a Deployment, StatefulSet, or ReplicaSet are in a ready state before marking the release as successful. It will wait for as long as --timeout")
	f.BoolVar(&client.Helm.Atomic, "atomic", false, "if set, upgrade process rolls back changes made in case of failed upgrade. The --wait flag will be set automatically if --atomic is used")
	f.IntVar(&client.Helm.MaxHistory, "history-max", 10, "limit the maximum number of revisions saved per release. Use 0 for no limit")
	//f.BoolVar(&client.Helm.CleanupOnFail, "cleanup-on-fail", false, "allow deletion of new resources created in this upgrade when upgrade fails")
	//f.BoolVar(&client.Helm.SubNotes, "render-subchart-notes", false, "if set, render subchart notes along with the parent")
	//f.StringVar(&client.Helm.Description, "description", "", "add a custom description")
	////addChartPathOptionsFlags(f, &client.ChartPathOptions)
	addValueOptionsFlags(f, valueOpts)
	bindOutputFlag(cmd, &outfmt)
	bindPostRenderFlag(cmd, &client.Helm.PostRenderer)

	origHelpFunc := cmd.HelpFunc()
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		hideHelmCommonFlags(cmd)
		origHelpFunc(cmd, args)
	})
	return cmd
}
