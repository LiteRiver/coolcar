package sim

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type Subscriber interface {
	Subscribe(ctx context.Context) (ch chan *carpb.CarEntity, cleanUp func(), err error)
}

type Controller struct {
	CarService carpb.CarServiceClient
	Subscriber Subscriber
	Logger     *zap.Logger
}

func (c *Controller) RunSimulations(ctx context.Context) {
	var cars []*carpb.CarEntity
	for {
		time.Sleep(3 * time.Second)
		res, err := c.CarService.GetCars(ctx, &carpb.GetCarsRequest{})
		if err != nil {
			c.Logger.Error("cannot get cars", zap.Error(err))
			continue
		}
		cars = res.Cars
		break
	}

	c.Logger.Info("Running car simulations.", zap.Int("car_count", len(cars)))

	msgChan, cleanUp, err := c.Subscriber.Subscribe(ctx)
	defer cleanUp()
	if err != nil {
		c.Logger.Error("cannot subscribe", zap.Error(err))
		return
	}

	carChans := make(map[string]chan *carpb.Car)
	for _, car := range cars {
		ch := make(chan *carpb.Car)
		carChans[car.Id] = ch
		go c.SimulateCar(context.Background(), car, ch)
	}

	for carUpdate := range msgChan {
		ch := carChans[carUpdate.Id]
		if ch != nil {
			ch <- carUpdate.Car
		}
	}
}

func (c *Controller) SimulateCar(ctx context.Context, initial *carpb.CarEntity, ch chan *carpb.Car) {
	carId := initial.Id
	c.Logger.Info("simulation running.", zap.String("car_id", carId))

	for update := range ch {
		fmt.Printf("receive: %+v\n", update)
		if update.Status == carpb.CarStatus_UNLOCKING {
			_, err := c.CarService.UpdateCar(ctx, &carpb.UpdateCarRequest{
				Id:     carId,
				Status: carpb.CarStatus_UNLOCKED,
			})

			if err != nil {
				c.Logger.Error("cannot unlock car", zap.Error(err))
			}
		} else if update.Status == carpb.CarStatus_LOCKING {
			_, err := c.CarService.UpdateCar(ctx, &carpb.UpdateCarRequest{
				Id:     carId,
				Status: carpb.CarStatus_LOCKED,
			})

			if err != nil {
				c.Logger.Error("cannot lock car", zap.Error(err))
			}
		}
	}
}
