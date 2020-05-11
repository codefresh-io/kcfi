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
	"log"
	"io"
	"github.com/spf13/viper"
	"github.com/spf13/cobra"

	"helm.sh/helm/v3/cmd/helm/require"
	"github.com/codefresh-io/kcfi/pkg/action"
)

const cfImagesDesc = `
Commmand to push images to private registry
Push whole release with images list defined in config file:
   kcfi images push-release [-c|--config /path/to/config.yaml] [options]

Push single image
  kcfi images push [-c|--config /path/to/config.yaml] [options] repo/image:tag [repo/image:tag] ... 
`

func cfImagesCmd(out io.Writer) *cobra.Command {
	var configFileName string
	
	cmd := &cobra.Command{
		Use:   "images",
		Short: "push images to private registry",
		Long:  cfImagesDesc,
		Aliases: []string{"image", "private-registry", "docker"},
		Args:  require.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			viper.SetConfigFile(configFileName)
			if err := viper.ReadInConfig(); err != nil {
				log.Fatal(err)
			}
			debug("Using config file: %s", viper.ConfigFileUsed())

			config := viper.AllSettings()
			imagesPusher, err := action.NewImagesPusherFromConfig(config)
			if err != nil {
				log.Fatal(err)
			}
			return imagesPusher.Run()
		},
	}

	f := cmd.Flags()
	f.StringVarP(&configFileName, flagConfig, "c", defaultConfigFileName(), "config file")
	
	return cmd
}