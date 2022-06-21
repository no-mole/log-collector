package logger_center

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/no-mole/neptune/grpc_pool"
	"github.com/no-mole/neptune/logger/dispatcher"
	"github.com/no-mole/neptune/logger/entry"
	"github.com/no-mole/neptune/logger/formatter"
	"github.com/no-mole/neptune/logger/tagger"
	loggerPb "github.com/no-mole/log-collector/protos/logger"
)

func init() {
	once = &sync.Once{}
	dispatcher.Registry("grpc", func(formatter formatter.Formatter, tagger tagger.Tagger, config *dispatcher.Config) dispatcher.Dispatcher {
		return NewGrpcDispatcher(config.GetString("tag", "grpc"), formatter, tagger)
	})
}

var once *sync.Once
var AlarmFrequency = 10 * time.Second
var MaxAlarmTimes = 16

func initGrpcPool(ctx context.Context) {
	option := grpc_pool.WithOptions(
		grpc_pool.WithMaxActive(2),
		grpc_pool.WithMaxIdle(1),
	)
	grpc_pool.Init(ctx,
		option,
		loggerPb.Metadata(),
	)
}

func NewGrpcDispatcher(tag string, formatter formatter.Formatter, tagger tagger.Tagger) dispatcher.Dispatcher {
	once.Do(func() {
		initGrpcPool(context.Background())
	})
	return &GrpcDispatcher{
		Helper:        dispatcher.NewHelper(formatter, tagger),
		tag:           tag,
		lastAlarmTime: time.Now().Add(-AlarmFrequency),
	}
}

type GrpcDispatcher struct {
	dispatcher.Helper
	tag           string
	lastAlarmTime time.Time
	alarmTimes    int
}

func (g *GrpcDispatcher) Dispatch(entries []entry.Entry) {
	conn, err := grpc_pool.GetConnection(loggerPb.Metadata())
	if err != nil {
		g.error("grpc dispatcher not available:%s", err.Error())
		return
	}
	defer conn.Close()

	cli := loggerPb.NewLoggerServiceClient(conn.Value())
	stream, err := cli.Write(context.Background())
	if err != nil {
		g.error("grpc dispatcher write error:%s", err.Error())
		return
	}

	for _, e := range entries {
		if !g.Match(e.GetTag()) {
			continue
		}
		err = stream.Send(&loggerPb.WriteRequest{
			Tag:   g.tag,
			Entry: g.Format(e),
		})
		if err != nil {
			g.error("grpc dispatcher stream send error:%s", err.Error())
			return
		}
	}

	_, err = stream.CloseAndRecv()
	if err != nil {
		return
	}
}

func (g *GrpcDispatcher) error(format string, args ...interface{}) {
	if g.alarmTimes > MaxAlarmTimes {
		return
	}
	if !g.lastAlarmTime.Before(time.Now().Add(-AlarmFrequency)) {
		return
	}
	g.alarmTimes++
	g.lastAlarmTime = time.Now()
	fmt.Printf(format, args...)
	fmt.Print("\n")
}

var _ dispatcher.Dispatcher = &GrpcDispatcher{}
