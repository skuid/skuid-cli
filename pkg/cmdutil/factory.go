package cmdutil

import (
	"github.com/skuid/skuid-cli/pkg/logging"
)

type Factory struct {
	// TODO: This can be extended to support DI to address https://github.com/skuid/skuid-cli/issues/166
	//       For example, adding appVersion, httpClient, etc. can then be passed to commands
	//       which can transpose from factory to purpose specific options (e.g., DeployOptions,
	//       RetrieveOptions, etc.) which the packages can then use.  All in the name of testability.

	AppVersion string
	// This can become a full logger (or add a separate Logger) - see comments on logging.LogInformer interface
	LogConfig logging.LogInformer
	Commander CommandInformer
}

func NewFactory(appVersion string) *Factory {
	return &Factory{
		AppVersion: appVersion,
		LogConfig:  logging.NewLogConfig(),
		Commander:  NewCommander(),
	}
}
