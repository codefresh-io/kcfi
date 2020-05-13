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
	"os"
	//"io/ioutil"
	"path/filepath"
	//"github.com/pkg/errors"
	"github.com/spf13/cobra"
	//"sigs.k8s.io/yaml"
	
	"helm.sh/helm/v3/cmd/helm/require"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"github.com/codefresh-io/kcfi/pkg/action"
	c "github.com/codefresh-io/kcfi/pkg/config"
)

const cfImagesDesc = `
Commmand to push images to private registry
Push whole release with images list defined in config file or by --image-list parameter:
   kcfi images push [--images-list <images-list-file>] [-c|--config /path/to/config.yaml] [options]

Push single image
  kcfi images push [-c|--config /path/to/config.yaml] [options] repo/image:tag [repo/image:tag] ... 
`

func cfImagesCmd(out io.Writer) *cobra.Command {
	cmdPush := "push"

	var configFileName, imagesListFile, cfRegistrySecret string
	var registry, registryUsername, registryPassword string
		
	cmd := &cobra.Command{
		Use:   "images",
		Short: "push images to private registry",
		Long:  cfImagesDesc,
		Aliases: []string{"image", "private-registry", "docker"},
		Args:  require.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 || !(args[0] == cmdPush) {
				return cmd.Usage()
			}
			
			valueOpts := &values.Options{}
			var baseDir string
			if configFileName != "" {
				if fInfo, err := os.Stat(configFileName); err == nil && !fInfo.IsDir(){
					debug("Using config file: %s", configFileName)
					valueOpts.ValueFiles = []string{configFileName}
					baseDir = filepath.Dir(configFileName)
				} else {
					return fmt.Errorf("%s is not a valid file", configFileName)
				}
			}
			valueOpts.Values = append(valueOpts.Values, fmt.Sprintf("%s=%s", c.KeyBaseDir, baseDir))
			if cfRegistrySecret != "" {
				valueOpts.Values = append(valueOpts.Values, fmt.Sprintf("%s=%s", c.KeyImagesCodefreshRegistrySa, cfRegistrySecret))
			}
			if registry != "" {
				valueOpts.Values = append(valueOpts.Values, fmt.Sprintf("%s=%s", c.KeyImagesPrivateRegistryAddress, registry))
			}
			if registryUsername != "" {
				valueOpts.Values = append(valueOpts.Values, fmt.Sprintf("%s=%s", c.KeyImagesPrivateRegistryUsername, registryUsername))
			}
			if registryPassword != "" {
				valueOpts.Values = append(valueOpts.Values, fmt.Sprintf("%s=%s", c.KeyImagesPrivateRegistryPassword, registryPassword))
			}
			config, err := valueOpts.MergeValues(getter.Providers{})
			if err != nil {
				return err
			}

			imagesPusher, err := action.NewImagesPusherFromConfig(config)
			if err != nil {
				return err
			}

			var imagesList []string
			if len(args[1:]) > 0 {
				imagesList = args[1:]
			} else if imagesListFile != "" {
				imagesList, err = action.ReadListFile(imagesListFile)
				if err != nil {
					return err
				}
			} else {
				imagesList = imagesPusher.ImagesList
			}

			return imagesPusher.Run(imagesList)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&configFileName, flagConfig, "c", "", "config file")
	f.StringVar(&imagesListFile, "images-list", "", "file with list of images to push")
	f.StringVar(&cfRegistrySecret, "codefresh-registry-secret", "", "file with Codefresh registry secret (sa.json)")
	f.StringVar(&registry, "registry", "", "registry address")
	f.StringVar(&registryUsername, "user", "", "registry username")
	f.StringVar(&registryPassword, "password", "", "registry password")
	return cmd
}

			// configFileB, err := ioutil.ReadFile(configFileName)
			// if err != nil {
			// 	return errors.Wrap(err, fmt.Sprintf("cannot read %s", configFileName))
			// }
			// config := map[string]interface{}{}
			// err = yaml.Unmarshal(configFileB, &config)
			// if err != nil {
			// 	return err
			// }
			
			
			// config[c.KeyBaseDir] = baseDir