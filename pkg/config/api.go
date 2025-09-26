package config

import (
	"github.com/forkspacer/api-server/pkg/utils"
	"github.com/hashicorp/go-multierror"
)

type APIConfig struct {
	Dev     bool
	APIPort uint16
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

	return &APIConfig{
		Dev:     dev,
		APIPort: apiPort,
	}, errs
}
