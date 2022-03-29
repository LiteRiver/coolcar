package car

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/id"
)

type Manager struct {
}

func (c *Manager) Verify(ctx context.Context, carId id.CarId, location *rentalpb.Location) error {
	return nil
}

func (c *Manager) Unlock(ctx context.Context, carId id.CarId) error {
	return nil
}
