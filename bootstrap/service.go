package bootstrap

import (
	"context"

	"github.com/no-mole/neptune/registry"
	"github.com/no-mole/log-collector/protos/logger"
)

func RegistrationService(ctx context.Context) error {
	return registry.Registry(ctx,
		logger.Metadata(),
	)
}
