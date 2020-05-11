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
	//"github.com/stretchr/objx"
)

// ImagesPusher pusher of images
type ImagesPusher struct {
	RegistryCredentials map[string]interface{}
	ImagesList []string
}

func NewImagesPusherFromConfig(config map[string]interface{}) (*ImagesPusher, error) {
	return &ImagesPusher{
		RegistryCredentials: config,
		ImagesList: []string{},
	}, nil
}

func(o *ImagesPusher) Run() error {
	fmt.Printf("Running images pusher")
	return nil
} 
