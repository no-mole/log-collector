package data_store

import (
	"context"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
	"github.com/no-mole/neptune/logger"

	pb "github.com/no-mole/log-collector/protos/logger"
)

var _ Datastore = &FileDataStore{}

func NewFileDadaStore(ctx context.Context, queueSize int, fileName string, maxSize, maxBackups, maxAge int) Datastore {
	return &FileDataStore{
		ctx: ctx,
		ch:  make(chan *pb.WriteRequest, queueSize),
		writer: &lumberjack.Logger{
			Filename:   fileName,
			MaxSize:    maxSize,    //最大M数，超过则切割
			MaxBackups: maxBackups, //最大文件保留数，超过就删除最老的日志文件
			MaxAge:     maxAge,     //保存30天
			Compress:   false,      //是否压缩
		},
	}
}

type FileDataStore struct {
	ctx    context.Context
	ch     chan *pb.WriteRequest
	writer *lumberjack.Logger

	once sync.Once
}

func (f *FileDataStore) Add(req *pb.WriteRequest) {
	f.once.Do(func() {
		go f.flush()
	})
	f.ch <- req
}

func (f *FileDataStore) flush() {
	for {
		select {
		case <-f.ctx.Done():
			return
		case req := <-f.ch:
			_, err := f.writer.Write(append(req.Entry, '\n'))
			if err != nil {
				logger.Error(f.ctx, "file_data_store_flush", err)
			}
		}
	}
}
