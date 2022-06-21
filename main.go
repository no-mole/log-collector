package main

import (
	"context"

	_ "go.uber.org/automaxprocs"

	biogo "github.com/no-mole/neptune/app"
	"github.com/no-mole/neptune/config"
	"github.com/no-mole/log-collector/bootstrap"
)

func main() {
	ctx := context.Background()

	biogo.NewApp(ctx)

	biogo.AddHook(
		config.Init,          //初始化配置
		bootstrap.InitLogger, //初始化日志
		bootstrap.InitEs,
		bootstrap.InitGrpcServer,         //初始化grpc server
		bootstrap.InitServiceDispatchers, //启动收集器
	)
	biogo.AddDelayHook(
		bootstrap.RegistrationService, //最后启动服务注册
	)

	if err := biogo.Start(); err != nil {
		panic(err)
	}

	err := <-biogo.ErrorCh()
	biogo.Stop()
	panic(err)
}
