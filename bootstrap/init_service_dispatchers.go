package bootstrap

import (
	"context"
	"os"
	"path"

	"github.com/no-mole/log-collector/service/logger_center"
	"gopkg.in/yaml.v2"

	"github.com/no-mole/neptune/config"
)

func InitServiceDispatchers(ctx context.Context) error {
	body, err := os.ReadFile(path.Join(config.GlobalConfig.BasePath, "config", "driver.yml"))
	if err != nil {
		return err
	}
	conf := &logger_center.Config{}
	err = yaml.Unmarshal(body, conf)
	if err != nil {
		return err
	}
	if len(conf.Outputs) == 0 {
		return nil
	}
	return logger_center.InitDispatchers(ctx, conf)
}
