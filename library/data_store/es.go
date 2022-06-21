package data_store

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/no-mole/neptune/logger"

	"github.com/olivere/elastic/v7"
	pb "github.com/no-mole/log-collector/protos/logger"
)

var _ Datastore = &ESDatastore{}

func NewESDatastore(ctx context.Context, client *elastic.Client, chanSize int, flushFrequency time.Duration) Datastore {
	ds := &ESDatastore{
		ctx:     ctx,
		client:  client,
		ticker:  time.NewTicker(flushFrequency),
		flushCh: make(chan struct{}, 1),
		dataCh:  make(chan *pb.WriteRequest, chanSize),
		once:    sync.Once{},
	}
	return ds
}

type ESDatastore struct {
	ctx context.Context

	client *elastic.Client

	dataCh chan *pb.WriteRequest

	flushCh chan struct{}

	ticker *time.Ticker

	once sync.Once
}

func (ds *ESDatastore) timer() {
	for {
		select {
		case <-ds.ticker.C:
			ds.Flush()
		case <-ds.ctx.Done():
			ds.Flush()
		}
	}
}

func (ds *ESDatastore) Flush() {
	select {
	case ds.flushCh <- struct{}{}:
	default:
	}
}

func (ds *ESDatastore) check() {
	if len(ds.dataCh)/cap(ds.dataCh)*100 > 75 {
		ds.Flush()
	}
}

func (ds *ESDatastore) Add(data *pb.WriteRequest) {
	logger.Trace(ds.ctx, "es_data_store", logger.WithField("tag", data.Tag), logger.WithField("entry", string(data.Entry)))
	ds.once.Do(func() {
		go ds.timer()
		go ds.flush()
	})
	select {
	case ds.dataCh <- data:
		ds.check()
	default:
		ds.Flush()
		go func() {
			ds.dataCh <- data
		}()
	}
}

func (ds *ESDatastore) flush() {
	for {
		select {
		case <-ds.flushCh:
			bulkRequest := ds.client.Bulk()
			t := time.Now().Format("2006-01-02")
			data := make([]*pb.WriteRequest, 0, len(ds.dataCh))
			flag := true
			for flag && len(data) < cap(data) {
				select {
				case d := <-ds.dataCh:
					data = append(data, d)
				default:
					flag = false
				}
			}

			for _, d := range data {
				bulkRequest = bulkRequest.Add(elastic.NewBulkIndexRequest().Index(fmt.Sprintf("%s-%s", d.Tag, t)).Doc(string(d.Entry)))
			}
			resp, err := bulkRequest.Do(ds.ctx)
			//插入失败或者请求无响应
			if err != nil {
				logger.Error(ds.ctx, "app", err)
				continue
			}
			logger.Trace(ds.ctx, "es_data_store_flush", logger.WithField("took", resp.Took))
		}
	}
}
