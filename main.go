package main

import (
	"context"

	_ "go.uber.org/automaxprocs"

	neptune "github.com/no-mole/neptune/app"
	"github.com/no-mole/neptune/config"
	"github.com/no-mole/log-collector/bootstrap"
)

func main() {
	ctx := context.Background()

	neptune.NewApp(ctx)

	neptune.AddHook(
		config.Init,          //初始化配置
		bootstrap.InitLogger, //初始化日志
		bootstrap.InitEs,
		bootstrap.InitGrpcServer,         //初始化grpc server
		bootstrap.InitServiceDispatchers, //启动收集器
	)
	neptune.AddDelayHook(
		bootstrap.RegistrationService, //最后启动服务注册
	)

	if err := neptune.Start(); err != nil {
		panic(err)
	}

	err := <-neptune.ErrorCh()
	neptune.Stop()
	panic(err)
}
