package bootstrap

import (
	"context"

	"github.com/no-mole/neptune/redis"
	"github.com/no-mole/log-collector/model"

	"github.com/no-mole/neptune/config"
	"github.com/no-mole/neptune/config/center"
	"github.com/no-mole/neptune/elastic_search"
	"github.com/olivere/elastic/v7"
)

var esNames = []string{
	model.ElasticSearchLoggerCenter,
}

func InitEs(ctx context.Context) error {
	configCenterClient := config.GetClient()
	for _, esName := range esNames {
		conf, err := configCenterClient.Get(ctx, esName)
		if err != nil {
			return err
		}
		err = elastic_search.InitElasticSearch(
			esName,
			conf.GetValue(),
			elastic.SetGzip(true),
		)
		if err != nil {
			return err
		}
		// 监听修改
		configCenterClient.Watch(ctx, conf, func(item *center.Item) {
			redis.Init(item.Key, item.GetValue())
		})
	}
	return nil
}
