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
	"path/filepath"
	"io/ioutil"
	"github.com/pkg/errors"
	"github.com/stretchr/objx"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"

	c "github.com/codefresh-io/kcfi/pkg/config"
)

// ImagesPusher pusher of images
type ImagesPusher struct {
	// CfRegistryAuthConfig *authn.AuthConfig
	// DstRegistryAuthConfig *authn.AuthConfig
	Keychain authn.Keychain
	DstRegistry name.Registry
	ImagesList []string
}

// pusherKeychain implements authn.Keychain with the semantics of the standard Docker
// credential keychain.
type pusherKeychain struct{
	cfRegistry name.Registry
	cfRegistryAuthConfig *authn.AuthConfig
	dstRegistry name.Registry
	dstRegistryAuthConfig *authn.AuthConfig
}

func(k *pusherKeychain) Resolve(target authn.Resource) (authn.Authenticator, error) {

  var authenticator authn.Authenticator
  key := target.RegistryStr()
  switch {
  case key == name.DefaultRegistry:
	authenticator = authn.Anonymous
  case key == k.cfRegistry.RegistryStr():
	authenticator = authn.FromConfig(*k.cfRegistryAuthConfig)
  case key == k.dstRegistry.RegistryStr():
	authenticator = authn.FromConfig(*k.dstRegistryAuthConfig)
  default:
	authenticator = authn.Anonymous
  }

  return authenticator, nil

}

func NewImagesPusherFromConfig(config map[string]interface{}) (*ImagesPusher, error) {
	
	cfgX := objx.New(config)
	baseDir := cfgX.Get(c.KeyBaseDir).String()

	// get AuthConfig Codefresh Enterprise registry
	cfRegistrySaVal := cfgX.Get(c.KeyImagesCodefreshRegistrySa).Str("sa.json")
	cfRegistrySaPath := path.Join(filepath.Dir(baseDir), cfRegistrySaVal)
	cfRegistryPasswordB, err := ioutil.ReadFile(cfRegistrySaPath)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("cannot read %s", cfRegistrySaPath))
	}
	cfRegistry, _ := name.NewRegistry(c.CfRegistryAddr)
	cfRegistryAuthConfig := &authn.AuthConfig{
		Username: c.CfRegistryUsername,
		Password: string(cfRegistryPasswordB),
	}
	

	dstRegistryAddress := cfgX.Get(c.KeyImagesPrivateRegistryAddress).String()
	dstRegistry, err := name.NewRegistry(dstRegistryAddress)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid registry address %s", dstRegistryAddress)
	}	
	dstRegistryUsername := cfgX.Get(c.KeyImagesPrivateRegistryUsername).String()
	dstRegistryPassword := cfgX.Get(c.KeyImagesPrivateRegistryPassword).String()
	if len(dstRegistryAddress) == 0 || len(dstRegistryUsername) == 0 || len(dstRegistryPassword) == 0 {
		err = fmt.Errorf("missing private registry data: ")
		if len(dstRegistryAddress) == 0 {
			err = errors.Wrapf(err, "missing %s", c.KeyImagesPrivateRegistryAddress)
		}
		if len(dstRegistryUsername) == 0 {
			err = errors.Wrapf(err, "missing %s", c.KeyImagesPrivateRegistryUsername)
		}
		if len(dstRegistryPassword) == 0 {
			err = errors.Wrapf(err, "missing %s", c.KeyImagesPrivateRegistryPassword)
		}
		return nil, err
	}

	dstRegistryAuthConfig := &authn.AuthConfig{
		Username: dstRegistryUsername,
		Password: dstRegistryPassword,
	}

	keychain := &pusherKeychain{
		cfRegistry: cfRegistry,
		cfRegistryAuthConfig: cfRegistryAuthConfig,
		dstRegistry: dstRegistry,
		dstRegistryAuthConfig: dstRegistryAuthConfig,
	}

	return &ImagesPusher{
		DstRegistry: dstRegistry,
		Keychain: keychain,
	}, nil
}

func(o *ImagesPusher) Run() error {
	fmt.Printf("Running images pusher")
	return nil
} 
