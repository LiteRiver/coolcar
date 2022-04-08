package car

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/id"
	"fmt"
)

type Manager struct {
	CarService carpb.CarServiceClient
}

func (m *Manager) Verify(ctx context.Context, carId id.CarId, location *rentalpb.Location) error {
	car, err := m.CarService.GetCar(ctx, &carpb.GetCarRequest{Id: carId.String()})
	if err != nil {
		return fmt.Errorf("cannot get car: %v", err)
	}

	if car.Status != carpb.CarStatus_LOCKED {
		return fmt.Errorf("cannot unlock, car status is %v", car.Status)
	}

	return nil
}

func (m *Manager) Unlock(ctx context.Context, carId id.CarId, accountId id.AccountId, avatarUrl string, tripId id.TripId) error {
	_, err := m.CarService.UnlockCar(ctx, &carpb.UnlockCarRequest{
		Id: carId.String(),
		Driver: &carpb.Driver{
			Id:        accountId.String(),
			AvatarUrl: avatarUrl,
		},
		TripId: tripId.String(),
	})

	if err != nil {
		return fmt.Errorf("cannot unlock the car: %v", err)
	}

	return nil
}

func (m *Manager) Lock(ctx context.Context, carId id.CarId) error {
	_, err := m.CarService.LockCar(ctx, &carpb.LockCarRequest{
		Id: carId.String(),
	})
	if err != nil {
		return fmt.Errorf("cannot lock the car: %v", err)
	}

	return nil
}
