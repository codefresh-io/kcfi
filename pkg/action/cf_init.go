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
	"github.com/codefresh-io/onprem-operator/pkg/embeded/stage"
// 
)

const (
	// DefaultConfigFileName - 
	DefaultConfigFileName = "config.yaml"
)

// CfInit is an action to create Codefresh config stage directory
type CfInit struct {
	stageDir string
}

// NewCfInit creates object
func NewCfInit(stageDir string) *CfInit {
	return &CfInit{
		stageDir: stageDir,
	}
}

// Run the action
func (o *CfInit) Run() error {
	assetName := "codefresh"
	fmt.Printf("Creating stage directory in %s\n", path.Join(o.stageDir, assetName))
	return stage.RestoreAssets(o.stageDir, assetName)
}
