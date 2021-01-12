package charts

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"regexp"
	"strings"

	eCharts "github.com/codefresh-io/kcfi/pkg/embeded/charts"
)

func Load(chartName string) (*chart.Chart, error) {

	var chartBufferedFile []*loader.BufferedFile
	isArchivedChart, _ := regexp.MatchString(`^.*\.tgz$`, chartName)
	if isArchivedChart {
		chartData, err := eCharts.Asset(chartName)
		if err != nil {
			return nil, errors.Wrapf(err, "Cannot find chart data for %s", chartName)
		}

		chartBufferedFile, err = loader.LoadArchiveFiles(bytes.NewReader(chartData))
		if err != nil {
			return nil, errors.Wrapf(err, "Cannot load archived file for %s", chartName)
		}
	} else {
		for _, name := range eCharts.AssetNames() {
			fileNamePrefix := fmt.Sprintf("%s/", chartName)
			fileNameReplaceRe, err := regexp.Compile(`^(` + fileNamePrefix + `)(.*$)`)
			if err != nil {
				return nil, errors.Wrapf(err, "Wrong chart name %s", chartName)
			}
			var chartFileName string
			var chartFileData []byte
			if strings.HasPrefix(name, fileNamePrefix) {
				chartFileName = fileNameReplaceRe.ReplaceAllString(name, "$2")
				chartFileData, err = eCharts.Asset(name)
				if err != nil {
					return nil, errors.Wrapf(err, "Failed to load chart file %s", name)
				}
				chartBufferedFile = append(chartBufferedFile, &loader.BufferedFile{
					Name: chartFileName,
					Data: chartFileData,
				})
			}
		}
	}
	return loader.LoadFiles(chartBufferedFile)
}
