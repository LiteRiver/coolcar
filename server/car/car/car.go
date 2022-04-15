package car

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	"coolcar/car/car/dao"
	"coolcar/shared/id"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Publisher interface {
	Publish(context.Context, *carpb.CarEntity) error
}

type Service struct {
	Logger    *zap.Logger
	Mongo     *dao.Mongo
	Publisher Publisher
	carpb.UnimplementedCarServiceServer
}

func (s *Service) CreateCar(ctx context.Context, in *carpb.CreateCarRequest) (*carpb.CarEntity, error) {
	cr, err := s.Mongo.CreateCar(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &carpb.CarEntity{
		Id:  cr.Id.Hex(),
		Car: cr.Car,
	}, nil
}

func (s *Service) GetCar(ctx context.Context, in *carpb.GetCarRequest) (*carpb.Car, error) {
	cr, err := s.Mongo.GetCar(ctx, id.CarId(in.Id))
	if err != nil {
		s.Logger.Error("cannot get car", zap.Error(err))
		return nil, status.Error(codes.NotFound, "")
	}

	return cr.Car, nil
}

func (s *Service) GetCars(ctx context.Context, in *carpb.GetCarsRequest) (*carpb.GetCarsResponse, error) {
	rows, err := s.Mongo.GetCars(ctx)
	if err != nil {
		s.Logger.Error("cannot get cars", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	var cars []*carpb.CarEntity
	for _, c := range rows {
		cars = append(cars, &carpb.CarEntity{
			Id:  c.Id.Hex(),
			Car: c.Car,
		})
	}

	return &carpb.GetCarsResponse{
		Cars: cars,
	}, nil
}
func (s *Service) LockCar(ctx context.Context, in *carpb.LockCarRequest) (*carpb.LockCarResponse, error) {
	car, err := s.Mongo.UpdateCar(ctx, id.CarId(in.Id), carpb.CarStatus_UNLOCKED, &dao.CarUpdate{
		Status: carpb.CarStatus_LOCKING,
	})

	if err != nil {
		code := codes.Internal
		if err == mongo.ErrNoDocuments {
			code = codes.NotFound
		}

		return nil, status.Errorf(code, "cannot update: %v", err)
	}

	s.publish(ctx, car)
	return &carpb.LockCarResponse{}, nil
}

func (s *Service) UnlockCar(ctx context.Context, in *carpb.UnlockCarRequest) (*carpb.UnlockCarResponse, error) {
	car, err := s.Mongo.UpdateCar(
		ctx,
		id.CarId(in.Id),
		carpb.CarStatus_LOCKED,
		&dao.CarUpdate{
			Status:       carpb.CarStatus_UNLOCKING,
			Driver:       in.Driver,
			UpdateTripId: true,
			TripId:       id.TripId(in.TripId),
		})

	if err != nil {
		code := codes.Internal
		if err == mongo.ErrNoDocuments {
			code = codes.NotFound
		}
		return nil, status.Errorf(code, "cannot update: %v", err)
	}

	s.publish(ctx, car)

	return &carpb.UnlockCarResponse{}, nil
}
func (s *Service) UpdateCar(ctx context.Context, in *carpb.UpdateCarRequest) (*carpb.UpdateCarResponse, error) {
	update := &dao.CarUpdate{
		Status:   in.Status,
		Position: in.Position,
	}

	if in.Status == carpb.CarStatus_LOCKED {
		update.Driver = &carpb.Driver{}
		update.UpdateTripId = true
		update.TripId = id.TripId("")
	}

	car, err := s.Mongo.UpdateCar(ctx, id.CarId(in.Id), carpb.CarStatus_CS_NOT_SPECIFIED, update)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	s.publish(ctx, car)
	return &carpb.UpdateCarResponse{}, nil
}

func (s *Service) publish(c context.Context, car *dao.CarRow) {
	err := s.Publisher.Publish(c, &carpb.CarEntity{
		Id:  car.Id.Hex(),
		Car: car.Car,
	})
	if err != nil {
		s.Logger.Error("cannot publish", zap.Error(err))
	}
}
