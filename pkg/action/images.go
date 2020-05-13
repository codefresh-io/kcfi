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
	"regexp"
	"io/ioutil"
	"log"
	"os"
	"io"
	"bufio"
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
	cfRegistry, _ := name.NewRegistry(c.CfRegistryAddress)
	var cfRegistryAuthConfig *authn.AuthConfig
	cfRegistrySaVal := cfgX.Get(c.KeyImagesCodefreshRegistrySa).Str("")
	if cfRegistrySaVal != "" {
		cfRegistrySaPath := path.Join(baseDir, cfRegistrySaVal)
		cfRegistryPasswordB, err := ioutil.ReadFile(cfRegistrySaPath)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("cannot read %s", cfRegistrySaPath))
		}
		cfRegistryAuthConfig = &authn.AuthConfig{
			Username: c.CfRegistryUsername,
			Password: string(cfRegistryPasswordB),
		}
	} else {
		info("Warning: Codefresh registry credentials are not set")
		cfRegistryAuthConfig = &authn.AuthConfig{}
	}
	
	// get AuthConfig for destination provate registry
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

	// Get Images List
	var imagesListsFiles, imagesList []string
	// cfgX.Get(c.KeyImagesLists).StrSlice() - not working, returns empty
	imagesListsFilesI := cfgX.Get(c.KeyImagesLists).Data()
	if fileNamesI, ok := imagesListsFilesI.([]interface{}); ok {
		for _, f := range fileNamesI {
			if str, isStr := f.(string); isStr {
				imagesListsFiles = append(imagesListsFiles, str)
			} else {
				info("Warning: %s - %v is not a string", c.KeyImagesLists, f)
			}
		}
		debug("%v - %v", imagesListsFilesI, imagesListsFiles)
	} else if imagesListsFilesI != nil {
		info("Warning: %s - %v is not a list", c.KeyImagesLists, imagesListsFilesI)
	}

	for _, imagesListFile := range(imagesListsFiles) {
		imagesListF, err := ReadListFile(path.Join(baseDir, imagesListFile))
		if err != nil {
			info("Error: failed to read %s - %v", imagesListFile, err)
			continue
		}
		for _, image := range imagesListF {
			imagesList = append(imagesList, image)
		}
	}

	return &ImagesPusher{
		DstRegistry: dstRegistry,
		Keychain: keychain,
		ImagesList: imagesList,
	}, nil
}

func(o *ImagesPusher) Run(images []string) error {
	info("Running images pusher")
	if len(images) == 0 {
		info("No images to push")
		return nil
		// if len(o.ImagesList) == 0 {
		// 	info("No images to push")
		// 	return nil
		// }
		// images = o.ImagesList
	}
	imagesWarnings := make(map[string]string)

	for _, imgName := range images {
		info("\n------------------\nSource Image: %s", imgName)
		imgRef, err := name.ParseReference(imgName)
		if err != nil {
			imagesWarnings[imgName] = fmt.Sprintf("cannot parse %s - %v", imgName, err)
			info("Warning: %s", imagesWarnings[imgName])
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
			imagesWarnings[imgName] = fmt.Sprintf("cannot convert image name %s to destination image", imgName)
			info("Error: %s", imagesWarnings[imgName])
			continue
		}
		dstRef, err := name.ParseReference(dstImageName)
		if err != nil {
			imagesWarnings[imgName] = fmt.Sprintf("cannot parse %s - %v", dstImageName, err)
			info("Error: %s", imagesWarnings[imgName])
			continue
		}

		img, err := remote.Image(imgRef, remote.WithAuthFromKeychain(o.Keychain))
		if err != nil {
			imagesWarnings[imgName] = fmt.Sprintf("cannot get source image %s - %v", imgName, err)
			info("Error: %s", imagesWarnings[imgName])
			continue
		}

		info("Dest.  Image: %s", dstImageName)
		err = remote.Write(dstRef, img, remote.WithAuthFromKeychain(o.Keychain))
		if err != nil {
			imagesWarnings[imgName] = fmt.Sprintf("failed  %s to %s - %v", imgName, dstImageName, err)
			info("Error: %s", imagesWarnings[imgName])
			continue
		}
	}
	
	cntProcessed := len(images)
	cnfFail := len(imagesWarnings)
	cntSucess := cntProcessed - cnfFail
	if len(imagesWarnings) > 0 {
		info("\n----- %d images were failed:", cnfFail)
		for img, errMsg := range imagesWarnings {
			info("%s - %s", img, errMsg)
		}
	}
	info("\n----- Completed! -----\n%d of %d images were successfully pushed", cntSucess, cntProcessed)

	return nil
} 

// ReadListFile - reads file and returns list of strings with strimmed lines without #-comments, empty lines
func ReadListFile(fileName string) ([]string, error) {
	debug("Reading List File %s", fileName)
	lines := []string{}
	file, err := os.Open(fileName)
	defer file.Close()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file %s", fileName)
	}
	reader := bufio.NewReader(file)

	commentLineRe, _ := regexp.Compile(`^ *#+.*$`)
	nonEmptyLineRe, _ := regexp.Compile(`[a-zA-Z0-9]`)
	for {
		lineB, prefix, err :=  reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, errors.Wrapf(err, "failed to read file %s", fileName)
			}
		}
		if prefix {
			info("Warning: too long lines in %s", fileName)
			continue
		}
		line := string(lineB)
		if commentLineRe.MatchString(line) || ! nonEmptyLineRe.MatchString(line) {
			continue
		}
		lines = append(lines, strings.Trim(line, " "))
	}
	if len(lines) == 0 {
		info("Warning: no valid lines in file %s", fileName)
	}
	return lines, nil
}