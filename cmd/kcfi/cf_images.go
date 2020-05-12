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
	"io/ioutil"
	"path/filepath"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"helm.sh/helm/v3/cmd/helm/require"
	"github.com/codefresh-io/kcfi/pkg/action"
	c "github.com/codefresh-io/kcfi/pkg/config"
)

const cfImagesDesc = `
Commmand to push images to private registry
Push whole release with images list defined in config file:
   kcfi images push-list [--images-list <images-list-file>] [-c|--config /path/to/config.yaml] [options]

Push single image
  kcfi images push [-c|--config /path/to/config.yaml] [options] repo/image:tag [repo/image:tag] ... 
`

func cfImagesCmd(out io.Writer) *cobra.Command {
	cmdPush := "push"
	cmdPushList := "push-list"
	var configFileName string
	
	cmd := &cobra.Command{
		Use:   "images",
		Short: "push images to private registry",
		Long:  cfImagesDesc,
		Aliases: []string{"image", "private-registry", "docker"},
		Args:  require.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 || ! (args[0] == cmdPush || args[0] == cmdPushList) {
				return cmd.Usage()
			}
			// read config file
			debug("Using config file: %s", configFileName)
			configFileB, err := ioutil.ReadFile(configFileName)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("cannot read %s", configFileName))
			}
			config := map[string]interface{}{}
			err = yaml.Unmarshal(configFileB, &config)
			if err != nil {
				return err
			}
			
			baseDir := filepath.Dir(configFileName)
			config[c.KeyBaseDir] = baseDir

			imagesPusher, err := action.NewImagesPusherFromConfig(config)
			if err != nil {
				return err
			}

			var imagesList []string
			if args[0] == cmdPush {
				imagesList = args[1:]
			} else if args[0] == cmdPushList {
				imagesList = imagesPusher.ImagesList
			}
			return imagesPusher.Run(imagesList)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&configFileName, flagConfig, "c", defaultConfigFileName(), "config file")
	
	return cmd
}