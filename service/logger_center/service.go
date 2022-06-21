package logger_center

import (
	"context"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/no-mole/neptune/utils"

	"github.com/no-mole/neptune/elastic_search"
	"github.com/no-mole/log-collector/library/data_store"
	"github.com/no-mole/log-collector/model"
	pb "github.com/no-mole/log-collector/protos/logger"
)

type Service struct {
	pb.UnimplementedLoggerServiceServer
	ch chan *pb.WriteRequest
}

func New() pb.LoggerServiceServer {
	s := &Service{
		ch: make(chan *pb.WriteRequest, 1024),
	}
	go s.flush()
	return s
}

var dispatchers []*Dispatcher

func InitDispatchers(ctx context.Context, conf *Config) (err error) {
	dispatchers = make([]*Dispatcher, 0, len(conf.Outputs))
	logPath := utils.GetCurrentAbPath()
	for _, output := range conf.Outputs {
		var d data_store.Datastore
		switch output.Type {
		case "elasticsearch":
			client, exist := elastic_search.Client.GetClient(model.ElasticSearchLoggerCenter)
			if !exist {
				return fmt.Errorf("es client %s not found", model.ElasticSearchLoggerCenter)
			}

			d = data_store.NewESDatastore(
				ctx,
				client,
				int(output.GetInt64("queueSize", 1024)),
				time.Duration(output.GetInt64("flushFrequency", 5))*time.Second,
			)
		case "file":
			d = data_store.NewFileDadaStore(
				ctx,
				int(output.GetInt64("queueSize", 1024)),
				path.Join(logPath, "log", output.GetString("filename", output.Tags[0]+".d.log")),
				int(output.GetInt64("maxSize", 1024)),
				int(output.GetInt64("maxBackups", 10)),
				int(output.GetInt64("maxAge", 15)),
			)
		case "udp":
			if output.GetString("ip", "") == "" {
				err = errors.New("udp ip must be set")
				continue
			}
			if output.GetInt64("port", 0) == 0 {
				err = errors.New("udp ip must be set")
				continue
			}
			d, err = data_store.NewUdpDataStore(
				ctx,
				output.GetString("ip", ""),
				int(output.GetInt64("port", 0)),
				int(output.GetInt64("queueSize", 1024)),
			)
		default:
			return errors.New("not supported output type")
		}
		if err != nil {
			return err
		}
		tags := make(map[string]struct{}, len(output.Tags))
		for _, tag := range output.Tags {
			tags[tag] = struct{}{}
		}
		dispatchers = append(dispatchers, &Dispatcher{
			Tags: tags,
			ds:   d,
		})
	}
	return nil
}
