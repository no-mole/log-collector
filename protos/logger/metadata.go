package logger

import (
	"github.com/no-mole/neptune/registry"
)

func Metadata() *registry.Metadata {
	return &registry.Metadata{
		ServiceDesc: LoggerService_ServiceDesc,
		Namespace:   "neptune",
		Version:     "v1",
	}
}
