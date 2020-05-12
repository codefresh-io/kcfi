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
	"strings"
	"io/ioutil"
	"log"
	"os"
	"github.com/pkg/errors"
	"github.com/stretchr/objx"

	clogs "github.com/google/go-containerregistry/pkg/logs"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"

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


func init() {
	//initialize go-conteinerregistry logger
	clogs.Warn = log.New(os.Stderr, "", log.LstdFlags)
	clogs.Progress = log.New(os.Stdout, "", log.LstdFlags)

	if os.Getenv(c.EnvPusherDebug) != "" {
		clogs.Debug = log.New(os.Stderr, "", log.LstdFlags)
	}
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
	cfRegistrySaPath := path.Join(baseDir, cfRegistrySaVal)
	cfRegistryPasswordB, err := ioutil.ReadFile(cfRegistrySaPath)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("cannot read %s", cfRegistrySaPath))
	}
	cfRegistry, _ := name.NewRegistry(c.CfRegistryAddress)
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

func(o *ImagesPusher) Run(images []string) error {
	info("Running images pusher")
	for _, imgName := range images {
		info("\n------------------\nSource Image: %s", imgName)
		imgRef, err := name.ParseReference(imgName)
		if err != nil {
			info("Warning: cannot parse %s - %v", imgName, err)
			continue
		}

		// Calculating destination image
		/* there are 3 types of image names:
		# 1. non-codefresh like "bitnami/mongo:123" - convert to "private-registry-addr/bitnami/mongo:123"
		# 2. codefresh public images like "codefresh/engine:123" - convert to "private-registry-addr/codefresh/mongo:123"
		# 3. codefresh private images like gcr.io/codefresh-enterprise/codefresh/cf-api:cf-onprem-v1.0.86 - will be convert to "private-registry-addr/codefresh/mongo:123
		DELIMITER='codefresh/'
		*/
		var dstImageName string
		imgNameSplit := strings.Split(imgName, "codefresh/")
		if len(imgNameSplit) == 1 {
			dstImageName = fmt.Sprintf("%s/%s", o.DstRegistry.RegistryStr(), imgName)
		} else if len(imgNameSplit) == 2 {
			dstImageName = fmt.Sprintf("%s/codefresh/%s", o.DstRegistry.RegistryStr(), imgNameSplit[1])
		} else {
			info("Warning: cannot convert image name %s to destination image", imgName)
			continue
		}
		dstRef, err := name.ParseReference(dstImageName)
		if err != nil {
			info("Warning: cannot parse %s - %v", dstImageName, err)
			continue
		}

		img, err := remote.Image(imgRef, remote.WithAuthFromKeychain(o.Keychain))
		if err != nil {
			info("Warning: source image %s - %v", imgName, err)
			continue
		}

		info("Dest.  Image: %s", dstImageName)
		err = remote.Write(dstRef, img, remote.WithAuthFromKeychain(o.Keychain))
		if err != nil {
			info("Warning: failed  %s to %s - %v", imgName, dstImageName, err)
			continue
		}
	}
	return nil
} 
