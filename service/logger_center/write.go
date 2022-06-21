package logger_center

import (
	"context"
	"io"

	"github.com/no-mole/neptune/logger"

	"google.golang.org/protobuf/types/known/emptypb"
	pb "github.com/no-mole/log-collector/protos/logger"
)

func (s *Service) Write(svr pb.LoggerService_WriteServer) (err error) {
	var req *pb.WriteRequest
	for {
		req, err = svr.Recv()
		if err != nil {
			if err != io.EOF {
				logger.Error(context.Background(), "app", err)
			}
			break
		}
		s.ch <- req
		//ElasticSearchDatastore.Add(req)
		//FileDatastore.Add(req)
	}
	err = svr.SendAndClose(&emptypb.Empty{})
	return
}

func (s *Service) flush() {
	for {
		req := <-s.ch
		for _, dispatcher := range dispatchers {
			dispatcher.Dispatch(req)
		}
	}
}
