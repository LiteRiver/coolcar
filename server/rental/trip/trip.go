package trip

import (
	"context"
	"coolcar/rental/api/gen/v1"
	"coolcar/rental/trip/dao"
	"coolcar/shared/auth"
	"coolcar/shared/id"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	ProfileManager
	CarManager
	PointManager
	Logger *zap.Logger
	Mongo  *dao.Mongo
	rentalpb.UnimplementedTripServiceServer
}

// ACL(Anti Corruption Layer)
type ProfileManager interface {
	Verify(context.Context, id.AccountId) (id.IdentityId, error)
}

type CarManager interface {
	Verify(context.Context, id.CarId, *rentalpb.Location) error
	Unlock(context.Context, id.CarId) error
}

type PointManager interface {
	Resolve(context.Context, *rentalpb.Location) (string, error)
}

// TODO: 验证驾驶者身份
// TODO: 检查车辆状态
// TODO: 创建行程：写入数据库， 开始计费
// TODO: 车辆开锁
func (s *Service) CreateTrip(ctx context.Context, req *rentalpb.CreateTripRequest) (*rentalpb.TripEntity, error) {
	accountId, err := auth.AccountIdFromContext(ctx)
	if err != nil {
		return nil, err
	}

	identityId, err := s.ProfileManager.Verify(ctx, accountId)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	carId := id.CarId(req.CarId)
	err = s.CarManager.Verify(ctx, carId, req.Start)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	pointName, err := s.PointManager.Resolve(ctx, req.Start)
	if err != nil {
		s.Logger.Warn("cannot resolve point name", zap.Stringer("location", req.Start), zap.Error(err))
	}

	ls := &rentalpb.LocationStatus{
		Location:  req.Start,
		PointName: pointName,
	}
	trip, err := s.Mongo.CreateTrip(ctx, &rentalpb.Trip{
		AccountId:  accountId.String(),
		CarId:      carId.String(),
		IdentityId: identityId.String(),
		Status:     rentalpb.TripStatus_IN_PROGRESS,
		Start:      ls,
		Current:    ls,
	})

	if err != nil {
		s.Logger.Warn("cannot create trip", zap.Error(err))
		return nil, status.Error(codes.AlreadyExists, "")
	}

	go func() {
		err := s.CarManager.Unlock(context.Background(), carId)
		if err != nil {
			s.Logger.Error("cannot unlock car", zap.Error(err))
		}
	}()

	return &rentalpb.TripEntity{
		Id:   trip.Id.Hex(),
		Trip: trip.Trip,
	}, nil
}

func (s *Service) UpdateTrip(ctx context.Context, req *rentalpb.UpdateTripRequest) (*rentalpb.Trip, error) {
	accountId, err := auth.AccountIdFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "")
	}
	tripId := id.TripId(req.Id)
	row, err := s.Mongo.GetTrip(ctx, tripId, accountId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "")
	}

	if req.Current != nil {
		row.Trip.Current.Location = req.Current
		row.Trip.Current = s.calcCurrentStatus(row.Trip, req.Current)
	}

	if req.EndTrip {
		row.Trip.End = row.Trip.Current
		row.Trip.Status = rentalpb.TripStatus_FINISHED
	}

	s.Mongo.UpdateTrip(ctx, tripId, accountId, row.UpdatedAt, row.Trip)
	return row.Trip, nil
}

func (s *Service) calcCurrentStatus(trip *rentalpb.Trip, current *rentalpb.Location) *rentalpb.LocationStatus {
	return nil
}
