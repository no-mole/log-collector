package bootstrap

import (
	"context"

	"github.com/no-mole/log-collector/protos/logger"
	"github.com/no-mole/log-collector/service/logger_center"
	"github.com/no-mole/neptune/app"
)

func InitGrpcServer(_ context.Context) error {
	//不需要tracing
	s := app.NewGrpcServer()
	s.RegisterService(&logger.Metadata().ServiceDesc, logger_center.New())
	return nil
}
