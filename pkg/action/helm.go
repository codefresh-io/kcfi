package action

import (
	"fmt"
	"github.com/codefresh-io/kcfi/pkg/charts"
	c "github.com/codefresh-io/kcfi/pkg/config"
	"github.com/pkg/errors"
	"github.com/stretchr/objx"
	"os"
	"strings"
	"time"

	helm "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/output"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

type HelmChartOptions struct {
	ChartName string
	baseDir   string
	*helm.ChartPathOptions
}

func NewHelmChartOptionsFromConfig(chartName string, config map[string]interface{}) (*HelmChartOptions, error) {
	cfgX := objx.New(config)
	baseDir := cfgX.Get(c.KeyBaseDir).String()

	if chartName == "" {
		return nil, fmt.Errorf("Missing chart name in config")
	}
	helmChartOptions := &helm.ChartPathOptions{
		RepoURL: cfgX.Get(c.KeyHelmRepoURL).String(),
		Version: cfgX.Get(c.KeyHelmVersion).String(),

		Password: cfgX.Get(c.KeyHelmPassword).String(),
		Username: cfgX.Get(c.KeyHelmUsername).String(),
		Verify:   cfgX.Get(c.KeyHelmVerify).Bool(false),
		CaFile:   cfgX.Get(c.KeyHelmCaFile).String(),
		CertFile: cfgX.Get(c.KeyHelmCertFile).String(),
		KeyFile:  cfgX.Get(c.KeyHelmKeyFile).String(),
		Keyring:  cfgX.Get(c.KeyHelmKeyring).String(),
	}

	return &HelmChartOptions{
		ChartName:        chartName,
		baseDir:          baseDir,
		ChartPathOptions: helmChartOptions,
	}, nil
}

func (h *HelmChartOptions) LoadChart() (*chart.Chart, error) {

	settings := cli.New()
	settings.Debug = c.Debug
	if h.baseDir != "" {
		workingDir, _ := os.Getwd()
		os.Chdir(h.baseDir)
		defer os.Chdir(workingDir)
	}

	var ch *chart.Chart
	var err error
	if chartPath, err := h.ChartPathOptions.LocateChart(h.ChartName, settings); err == nil {
		debug("Using chart path %s", chartPath)
		ch, err = loader.Load(chartPath)
	} else {
		debug("Using embeded chart %s", h.ChartName)
		ch, err = charts.Load(h.ChartName)
	}
	return ch, err
}

// DeployHelmRelease - deploy helm release using chart from config
func DeployHelmRelease(releaseName string, chart string, vals map[string]interface{}, cfg *helm.Configuration, client *helm.Upgrade) (*release.Release, error) {

	info("Deploying release %s of chart %s ...", releaseName, chart)
	helmChartOptions, err := NewHelmChartOptionsFromConfig(chart, vals)
	if err != nil {
		return nil, err
	}
	chartRequested, err := helmChartOptions.LoadChart()
	if err != nil {
		return nil, err
	}
	if chartRequested == nil || chartRequested.Metadata == nil {
		return nil, fmt.Errorf("Failed to load %s chart. Check helm chart options in config", chart)
	}

	var release *release.Release

	// Checking if chart already installed and decide to use install or upgrade helm client
	histClient := helm.NewHistory(cfg)
	histClient.Max = 1
	if _, err := histClient.Run(releaseName); err == driver.ErrReleaseNotFound {
		// Only print this to stdout for table output

		info("Release %q does not exist. Installing it now.\n", releaseName)

		instClient := helm.NewInstall(cfg)
		instClient.CreateNamespace = false //TODO
		instClient.ChartPathOptions = client.ChartPathOptions
		instClient.DryRun = client.DryRun
		instClient.DisableHooks = client.DisableHooks
		instClient.SkipCRDs = client.SkipCRDs
		instClient.Timeout = client.Timeout
		instClient.Wait = client.Wait
		instClient.Devel = client.Devel
		instClient.Namespace = client.Namespace
		instClient.Atomic = client.Atomic
		instClient.PostRenderer = client.PostRenderer
		instClient.DisableOpenAPIValidation = client.DisableOpenAPIValidation
		instClient.SubNotes = client.SubNotes

		instClient.ReleaseName = releaseName

		if chartRequested.Metadata.Deprecated {
			fmt.Println("WARNING: This chart is deprecated")
		}
		return instClient.Run(chartRequested, vals)

	} else if err != nil {
		return nil, err
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		if err := helm.CheckDependencies(chartRequested, req); err != nil {
			return nil, err
		}
	}

	release, err = client.Run(releaseName, chartRequested, vals)
	if err != nil {
		return nil, errors.Wrapf(err, "UPGRADE of %s FAILED", releaseName)
	}
	info("Release %q has been upgraded\n", releaseName)

	return release, nil
}

// IsHelmReleaseInstalled - returns true if helm release installed
func IsHelmReleaseInstalled(releaseName string, cfg *helm.Configuration) bool {
	histClient := helm.NewHistory(cfg)
	histClient.Max = 1

	if release, err := histClient.Run(releaseName); release != nil {
		debug("release %s is installed", releaseName)
		return true
	} else {
		debug("query release %s returned error %v", releaseName, err)
		return false
	}
}

// PrintHelmReleaseInfo - prints helm release
func PrintHelmReleaseInfo(release *release.Release, debug bool) error {
	if release == nil {
		return nil
	}
	info("NAME: %s", release.Name)
	if !release.Info.LastDeployed.IsZero() {
		info("LAST DEPLOYED: %s", release.Info.LastDeployed.Format(time.ANSIC))
	}
	info("NAMESPACE: %s", release.Namespace)
	info("STATUS: %s", release.Info.Status.String())
	info("REVISION: %d", release.Version)

	out := os.Stdout
	if debug {
		info("USER-SUPPLIED VALUES:")
		err := output.EncodeYAML(out, release.Config)
		if err != nil {
			return err
		}
		// Print an extra newline
		fmt.Fprintln(out)

		cfg, err := chartutil.CoalesceValues(release.Chart, release.Config)
		if err != nil {
			return err
		}

		fmt.Fprintln(out, "COMPUTED VALUES:")
		err = output.EncodeYAML(out, cfg.AsMap())
		if err != nil {
			return err
		}
		// Print an extra newline
		fmt.Fprintln(out)
	}

	if strings.EqualFold(release.Info.Description, "Dry run complete") || debug {
		fmt.Fprintln(out, "HOOKS:")
		for _, h := range release.Hooks {
			fmt.Fprintf(out, "---\n# Source: %s\n%s\n", h.Path, h.Manifest)
		}
		fmt.Fprintf(out, "MANIFEST:\n%s\n", release.Manifest)
	}

	if len(release.Info.Notes) > 0 {
		fmt.Fprintf(out, "NOTES:\n%s\n", strings.TrimSpace(release.Info.Notes))
	}
	return nil
}

func GetReleaseValues(releaseName string, cfg *helm.Configuration) (map[string]interface{}, error) {
	debug("Getting values from the release named \"%s\"", releaseName)
	client := helm.NewGetValues(cfg)
	client.AllValues = true
	return client.Run(releaseName)
}
