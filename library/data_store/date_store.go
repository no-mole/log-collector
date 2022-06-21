package data_store

import pb "github.com/no-mole/log-collector/protos/logger"

type Datastore interface {
	Add(*pb.WriteRequest)
}
