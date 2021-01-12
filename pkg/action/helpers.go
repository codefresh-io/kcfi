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
	"bytes"
	"io/ioutil"
	"text/template"

	"github.com/codefresh-io/kcfi/pkg/engine"
	"sigs.k8s.io/yaml"
)

// ReadYamlFile - reads yaml file
func ReadYamlFile(fileName string) (map[string]interface{}, error) {
	fileB, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	var yamlResult map[string]interface{}
	if err := yaml.Unmarshal(fileB, &yamlResult); err != nil {
		return nil, err
	}

	return yamlResult, nil
}

// ExecuteTemplate - executes templates in tpl str with config as values
func ExecuteTemplate(tplStr string, data interface{}) (string, error) {

	template, err := template.New("base").Funcs(engine.FuncMap()).Parse(tplStr)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBufferString("")
	err = template.Execute(buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ExecuteTemplateToValues - exacutes template to map[string]interface{}
func ExecuteTemplateToValues(tplStr string, data interface{}) (map[string]interface{}, error) {
	tplResultS, err := ExecuteTemplate(tplStr, data)
	if err != nil {
		return nil, err
	}
	tplResult := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(tplResultS), &tplResult)
	return tplResult, err
}

// MergeMaps - merges two map[string]interface{} into one
func MergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = MergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
