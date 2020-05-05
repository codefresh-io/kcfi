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

	"io"
	"os"

	"github.com/spf13/cobra"

	"helm.sh/helm/v3/cmd/helm/require"
	"github.com/codefresh-io/kcfi/pkg/action"
)

const cfInitDesc = `
This command initializes installer by creating staging directory for Codefresh configuration
   kcfi init [/path/to/codefresh-config-dir]
by default creates in current directory
`

func cfInitCmd(out io.Writer) *cobra.Command {
	
	cmd := &cobra.Command{
		Use:   "init",
		Short: "initialize stage config directory",
		Long:  cfInitDesc,
		Args:  require.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var stageDir string
			var err error
			if len(args) > 0 {
				stageDir = args[0]
			} else {
				stageDir, err = os.Getwd()
				if err != nil {
					return err
				}
			}
			client := action.NewCfInit(stageDir)
			return client.Run()
		},
	}
	return cmd
}