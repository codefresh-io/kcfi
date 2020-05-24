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

package config

import (
	"os"
	"strconv"
)

const (
	// AssetsDir - folder name where we save kubernetes and helm assets
	AssetsDir = "assets"

	CodefreshReleaseName = "cf"
	OperatorReleaseName = "cf-onprem-operator"

	CfRegistryAddress = "gcr.io"
	CfRegistryUsername = "_json_key"

	KeyKind = "metadata.kind"
	KindCodefresh = "codefresh"
	KindK8sAgent = "k8sAgent"
	KindVenona = "venona"
	
	KeyImagesCodefreshRegistrySa = "images.codefreshRegistrySa"
	KeyImagesUsePrivateRegistry = "images.usePrivateRegistry"
	KeyImagesPrivateRegistryAddress = "images.privateRegistry.address"
	KeyImagesPrivateRegistryUsername = "images.privateRegistry.username"
	KeyImagesPrivateRegistryPassword = "images.privateRegistry.password"
	KeyImagesLists = "images.lists"

	KeyRelease = "metadata.installer.release"
	KeyInstallerType = "metadata.installer.type"
	InstallerTypeOperator = "operator"
	InstallerTypeHelm = "helm"

	KeyOperatorChartValues = "metadata.installer.operator"
	KeyOperatorSkipCRD = "metadata.installer.operator.skipCRD"
	KeyDockerRegistry = "dockerRegistry"

	OperatorHelmReleaseName = "cf-onprem-operator"
	OperatorHelmChartName = "codefresh-operator"
	
	KeyHelmChart = "metadata.installer.helm.chart"
	KeyHelmRelease = "metadata.installer.helm.release"
	KeyHelmRepoURL = "metadata.installer.helm.repoUrl"
	KeyHelmVersion = "metadata.installer.helm.version"
	KeyHelmCaFile = "metadata.installer.helm.caFile"
	KeyHelmCertFile = "metadata.installer.helm.certFile"
	KeyHelmKeyFile = "metadata.installer.helm.keyFile"
	KeyHelmKeyring = "metadata.installer.helm.keyring"
	KeyHelmPassword = "metadata.installer.helm.password"
	KeyHelmUsername = "metadata.installer.helm.username"
	KeyHelmVerify = "metadata.installer.helm.verify"

	CodefreshHelmReleaseName = "cf"

	KeyKubeNamespace = "kubernetes.namespace"
	KeyKubeContext = "kubernetes.context"
	KeyKubeKubeconfig = "kubernetes.kubeconfig"

	KeyBaseDir = "BaseDir"
	KeyTlsSelfSigned = "tls.selfSigned"
	KeyTlsCert = "tls.cert"
	KeyTlsKey = "tls.key"

	KeyAppUrl = "global.appUrl"

	KeyDbInfra = "dbinfra"
	KeyDbInfraEnabled = "dbinfra.enabled"
	KeyDbInfraUpgrade = "dbinfra.upgrade"
	
	DbInfraHelmChartName = "codefresh-db-infra"
	DbInfraMainConfigChangeValuesFile = "values/db-infra.yaml"
	DbInfraConfigFile = "addons/db-infra/config.yaml"

	KeyInclude = "include"

	EnvPusherDebug = "PUSHER_DEBUG"
)

var Debug, _ = strconv.ParseBool(os.Getenv("HELM_DEBUG"))