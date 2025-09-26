package config

import (
	"os"
	"path/filepath"

	"github.com/forkspacer/api-server/pkg/utils"
	"github.com/hashicorp/go-multierror"
)

type KubernetesConfig struct {
	KubeFile string
}

type APIConfig struct {
	Dev     bool
	APIPort uint16
	KubernetesConfig
}

func NewAPIConfig() (*APIConfig, *multierror.Error) {
	var errs *multierror.Error

	dev, err := utils.GetEnvOr("DEV", true)
	if err != nil && err != utils.ErrEnvNotFound {
		errs = multierror.Append(err, errs)
	}

	apiPort, err := utils.GetEnvOr[uint16]("API_PORT", 8421)
	if err != nil && err != utils.ErrEnvNotFound {
		errs = multierror.Append(err, errs)
	}

	var kubernetesConfigKubeFile string
	if homeDir, err := os.UserHomeDir(); err != nil {
		errs = multierror.Append(multierror.Prefix(err, "K8S_KUBE_FILE"), errs)
	} else {
		kubernetesConfigKubeFile, err = utils.GetEnvOr[string]("K8S_KUBE_FILE", filepath.Join(homeDir, ".kube", "config"))
		if err != nil && err != utils.ErrEnvNotFound {
			errs = multierror.Append(err, errs)
		}
	}

	return &APIConfig{
		Dev:     dev,
		APIPort: apiPort,
		KubernetesConfig: KubernetesConfig{
			KubeFile: kubernetesConfigKubeFile,
		},
	}, errs
}
