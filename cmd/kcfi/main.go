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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"k8s.io/klog"
	"sigs.k8s.io/yaml"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/gates"
	"helm.sh/helm/v3/pkg/kube"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"

	c "github.com/codefresh-io/kcfi/pkg/config"
)

// FeatureGateOCI is the feature gate for checking if `helm chart` and `helm registry` commands should work
const FeatureGateOCI = gates.Gate("HELM_EXPERIMENTAL_OCI")

var (
	settings         *cli.EnvSettings
	defaultNamespace = "codefresh"

	flagConfig = "config"
)

var configuredNamespace string

func init() {
	log.SetFlags(log.Lshortfile)
	//Set Codefresh default namespace
	// if _, ok := os.LookupEnv("HELM_NAMESPACE"); !ok {
	// 	os.Setenv("HELM_NAMESPACE", defaultNamespace)
	// }
	settings = cli.New()
}

func debug(format string, v ...interface{}) {
	if settings.Debug {
		format = fmt.Sprintf("[debug] %s\n", format)
		log.Output(2, fmt.Sprintf(format, v...))
	}
}

func initKubeLogs() {
	pflag.CommandLine.SetNormalizeFunc(wordSepNormalizeFunc)
	gofs := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(gofs)
	pflag.CommandLine.AddGoFlagSet(gofs)
	pflag.CommandLine.Set("logtostderr", "true")
}

func main() {
	initKubeLogs()

	actionConfig := new(action.Configuration)
	cmd := newRootCmd(actionConfig, os.Stdout, os.Args[1:])

	configuredNamespace = settings.Namespace()
	// setting kubernetes client namespace, kube-context, kubeconfig
	// priorities:
	//   1. use flags --namespace, --kube-context, --kubeconfig. they are already set in o.Helm.*
	//   2. use values specified in config.yaml: {kubernetes: {namespace: "", kube-context: "", kubectonfig}}
	//   3. default context and namespace defined in ~/.kube/config
	// finding the command being executing and parse its config file to set kube client
	// we should do it here in main() because helm sets kube client here
	childCmd, childArgs, _ := cmd.Find(os.Args[1:])
	childFlags := childCmd.Flags()
	configFlag := childFlags.Lookup(flagConfig)
	if configFlag != nil {
		//merging config file kubernetes parameters into settings
		childFlags.Parse(childArgs)
		configFileName := configFlag.Value.String()
		if configFileName != "" {
			viper.SetConfigFile(configFileName)
			if err := viper.ReadInConfig(); err == nil {
				debug("Using config file: %s", viper.ConfigFileUsed())

				viper.BindPFlag(c.KeyKubeNamespace, childFlags.Lookup("namespace"))
				viper.BindPFlag(c.KeyKubeContext, childFlags.Lookup("kube-context"))
				viper.BindPFlag(c.KeyKubeKubeconfig, childFlags.Lookup("kubeconfig"))

				if ns := viper.GetString(c.KeyKubeNamespace); ns != "" {
					configuredNamespace = ns
				}
				if kubeContext := viper.GetString(c.KeyKubeContext); kubeContext != "" {
					settings.KubeContext = kubeContext
				}
				if kubeconfig := viper.GetString(c.KeyKubeKubeconfig); kubeconfig != "" {
					settings.KubeConfig = kubeconfig
				}
			}
		}
	}

	kubeClientConfig := kube.GetConfig(settings.KubeConfig, settings.KubeContext, configuredNamespace)
	if settings.KubeToken != "" {
		kubeClientConfig.BearerToken = &settings.KubeToken
	}
	if settings.KubeAPIServer != "" {
		kubeClientConfig.APIServer = &settings.KubeAPIServer
	}

	helmDriver := os.Getenv("HELM_DRIVER")
	debug("Initializing Kubernetes client: namespace=%s kube-context=%s kubeconfig =%s", configuredNamespace, settings.KubeContext, settings.KubeConfig)
	if err := actionConfig.Init(kubeClientConfig, configuredNamespace, helmDriver, debug); err != nil {
		log.Fatal(err)
	}
	if helmDriver == "memory" {
		loadReleasesInMemory(actionConfig)
	}

	if err := cmd.Execute(); err != nil {
		debug("%+v", err)
		switch e := err.(type) {
		case pluginError:
			os.Exit(e.code)
		default:
			os.Exit(1)
		}
	}
}

// wordSepNormalizeFunc changes all flags that contain "_" separators
func wordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	return pflag.NormalizedName(strings.ReplaceAll(name, "_", "-"))
}

func checkOCIFeatureGate() func(_ *cobra.Command, _ []string) error {
	return func(_ *cobra.Command, _ []string) error {
		if !FeatureGateOCI.IsEnabled() {
			return FeatureGateOCI.Error()
		}
		return nil
	}
}

// This function loads releases into the memory storage if the
// environment variable is properly set.
func loadReleasesInMemory(actionConfig *action.Configuration) {
	filePaths := strings.Split(os.Getenv("HELM_MEMORY_DRIVER_DATA"), ":")
	if len(filePaths) == 0 {
		return
	}

	store := actionConfig.Releases
	mem, ok := store.Driver.(*driver.Memory)
	if !ok {
		// For an unexpected reason we are not dealing with the memory storage driver.
		return
	}

	actionConfig.KubeClient = &kubefake.PrintingKubeClient{Out: ioutil.Discard}

	for _, path := range filePaths {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal("Unable to read memory driver data", err)
		}

		releases := []*release.Release{}
		if err := yaml.Unmarshal(b, &releases); err != nil {
			log.Fatal("Unable to unmarshal memory driver data: ", err)
		}

		for _, rel := range releases {
			if err := store.Create(rel); err != nil {
				log.Fatal(err)
			}
		}
	}
	// Must reset namespace to the proper one
	mem.SetNamespace(settings.Namespace())
}
