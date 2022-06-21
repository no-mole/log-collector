package logger_center

import (
	"github.com/no-mole/log-collector/library/data_store"
	pb "github.com/no-mole/log-collector/protos/logger"
)

type Dispatcher struct {
	Tags map[string]struct{} `json:"tags"`
	ds   data_store.Datastore
}

func (d *Dispatcher) Dispatch(req *pb.WriteRequest) {
	if _, ok := d.Tags[req.Tag]; ok {
		d.ds.Add(req)
	}
}
