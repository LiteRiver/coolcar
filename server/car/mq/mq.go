package mq

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
)

type Publisher interface {
	Publish(context.Context, *carpb.CarEntity) error
}

type Subscriber interface {
	Subscribe(ctx context.Context) (ch chan *carpb.CarEntity, cleanUp func(), err error)
}
