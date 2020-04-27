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

	"github.com/spf13/cobra"

	"helm.sh/helm/v3/cmd/helm/require"
	//"github.com/codefresh-io/onprem-operator/pkg/helm-internal/completion"
	
	"helm.sh/helm/v3/pkg/action"
	cAction "github.com/codefresh-io/onprem-operator/pkg/action"
)

const operatorDesc = `
This command controls Codefresh Operator
`

const deployDesc = `
This command deploys Codefresh operator
`

func newOperatorCmd(cfg *action.Configuration, out io.Writer) *cobra.Command {

	operatorCommand := &cobra.Command{
		Use:     "operator",
		Short:   "controls Codefresh Operator",
		Long:    operatorDesc,
		Args:    require.NoArgs,
	}

	deploySubCmd := &cobra.Command{
		Use:     "deploy",
		Short:   "deploys Codefresh Operator",
		Long:    deployDesc,
		Args:    require.NoArgs,
		RunE:    func(cmd *cobra.Command, args []string) error {
			operator := cAction.NewOperator()
			output, _ := operator.Deploy()
			fmt.Fprint(out, output)
			return nil
		},
	}

	operatorCommand.AddCommand(deploySubCmd)
	return operatorCommand
}