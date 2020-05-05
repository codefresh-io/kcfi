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
	"fmt"
	"strings"
	"github.com/spf13/cobra"

	"helm.sh/helm/v3/cmd/helm/require"
	"github.com/codefresh-io/kcfi/pkg/action"
)



func cfInitDesc() string {
	cfInitDesc := `
This command initializes installer by creating stage directory for Codefresh product configuration

  kcfi init <product> [-d /path/to/stage-dir]
	
products are: `
    cfInitDesc += strings.Join(action.StageDirsList(), ", ")
	cfInitDesc += `
by default it creates stage in current directory
`
	return cfInitDesc
}

func cfInitUse() string {
	cfInitUse := fmt.Sprintf("init %s [ -d /path/to/stagedir]", strings.Join(action.StageDirsList(), "|"))
	return cfInitUse
}

func cfInitCmd(out io.Writer) *cobra.Command {
	var productName, stageDir string
	cmd := &cobra.Command{
		Use:   cfInitUse(),
		Short: "initialize stage config directory",
		Long:  cfInitDesc(),
		Args:  require.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			productName = args[0]
			client := action.NewCfInit(productName, stageDir)
			return client.Run()
		},
	}

	f := cmd.Flags()
	f.StringVarP(&stageDir, "stage-dir", "d", "", "Codefresh config file")	
	return cmd
}