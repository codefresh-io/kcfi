package action

import (
	"fmt"
	"os"
	"time"
	"strings"
	"github.com/pkg/errors"
	helm "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"github.com/codefresh-io/kcfi/pkg/charts"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/storage/driver"
	"helm.sh/helm/v3/pkg/cli/output"
)

func DeployHelmRelease(releaseName string, chart string, vals map[string]interface{}, cfg *helm.Configuration, client *helm.Upgrade) (*release.Release, error) {
	var release *release.Release
	info("Deploying release %s of chart %s ...", releaseName, chart)
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
	
		// Check chart dependencies to make sure all are present in /charts
		chartRequested, err := charts.Load(chart)
		if err != nil {
			return nil, err
		}
		instClient.ReleaseName = releaseName
	
		if chartRequested.Metadata.Deprecated {
			fmt.Println("WARNING: This chart is deprecated")
		}
		return instClient.Run(chartRequested, vals)
	
	} else if err != nil {
		return nil, err
	}

	// Check chart dependencies to make sure all are present in /charts
	ch, err := charts.Load(chart)
	if err != nil {
		return nil, err
	}

	if req := ch.Metadata.Dependencies; req != nil {
		if err := helm.CheckDependencies(ch, req); err != nil {
			return nil, err
		}
	}

	release, err = client.Run(releaseName, ch, vals)
	if err != nil {
		return nil, errors.Wrapf(err, "UPGRADE of %s FAILED", releaseName)
	}
	info("Release %q has been upgraded\n", releaseName)

	return release, nil
}


// PrintHelmReleaseInfo - prints helm release
func PrintHelmReleaseInfo(release *release.Release, debug bool) error {
	if release == nil {
		return nil
	}
	info( "NAME: %s\n", release.Name)
	if !release.Info.LastDeployed.IsZero() {
		info( "LAST DEPLOYED: %s\n", release.Info.LastDeployed.Format(time.ANSIC))
	}
	info( "NAMESPACE: %s\n", release.Namespace)
	info( "STATUS: %s\n", release.Info.Status.String())
	info( "REVISION: %d\n", release.Version)


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
