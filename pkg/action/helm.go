package action

import (
	"fmt"
	"github.com/pkg/errors"
	helm "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"github.com/codefresh-io/onprem-operator/pkg/charts"
	"helm.sh/helm/v3/pkg/storage/driver"
)

func DeployHelmRelease(releaseName string, chart string, vals map[string]interface{}, cfg *helm.Configuration, client *helm.Upgrade) (*release.Release, error) {
	var release *release.Release
	fmt.Printf("Deploying release %s of chart %s ...", releaseName, chart)
	// Checking if chart already installed and decide to use install or upgrade helm client
	histClient := helm.NewHistory(cfg)
	histClient.Max = 1
	if _, err := histClient.Run(releaseName); err == driver.ErrReleaseNotFound {
		// Only print this to stdout for table output
	
		fmt.Printf("Release %q does not exist. Installing it now.\n", releaseName)
		
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
	fmt.Printf("Release %q has been upgraded\n", releaseName)

	return release, nil 	
}