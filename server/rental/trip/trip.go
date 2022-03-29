package trip

import (
	"context"
	"coolcar/rental/api/gen/v1"
	"coolcar/rental/trip/dao"
	"coolcar/shared/auth"
	"coolcar/shared/id"
	"math/rand"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const centsPerSec = 0.7

type Service struct {
	ProfileManager
	CarManager
	PointManager
	DistanceCalc
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

type DistanceCalc interface {
	DistanceKm(context.Context, *rentalpb.Location, *rentalpb.Location) (float64, error)
}

var nowFunc = func() int64 {
	return time.Now().Unix()
}

func (s *Service) GetTrip(ctx context.Context, req *rentalpb.GetTripRequest) (*rentalpb.Trip, error) {
	accountId, err := auth.AccountIdFromContext(ctx)
	if err != nil {
		return nil, err
	}

	row, err := s.Mongo.GetTrip(ctx, id.TripId(req.Id), accountId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "")
	}

	return row.Trip, nil
}

func (s *Service) GetTrips(ctx context.Context, req *rentalpb.GetTripsRequest) (*rentalpb.GetTripsResponse, error) {
	accountId, err := auth.AccountIdFromContext(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := s.Mongo.GetTrips(ctx, accountId, req.Status)
	if err != nil {
		s.Logger.Error("cannot get trips: %v", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	res := &rentalpb.GetTripsResponse{}
	for _, row := range rows {
		res.Trips = append(res.Trips, &rentalpb.TripEntity{
			Id:   row.Id.Hex(),
			Trip: row.Trip,
		})
	}

	return res, nil
}

func (s *Service) CreateTrip(ctx context.Context, req *rentalpb.CreateTripRequest) (*rentalpb.TripEntity, error) {
	accountId, err := auth.AccountIdFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.CarId == "" || req.Start == nil {
		return nil, status.Error(codes.InvalidArgument, "")
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

	ls := s.calcCurrentStatus(ctx, &rentalpb.LocationStatus{
		Location:     req.Start,
		TimestampSec: nowFunc(),
	}, req.Start)

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
		return nil, err
	}

	tripId := id.TripId(req.Id)
	row, err := s.Mongo.GetTrip(ctx, tripId, accountId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "")
	}

	if row.Trip.Status == rentalpb.TripStatus_FINISHED {
		return nil, status.Error(codes.FailedPrecondition, "cannot update after finished trip")
	}

	if row.Trip.Current == nil {
		s.Logger.Error("trip without Current set", zap.String("id", tripId.String()))
		return nil, status.Error(codes.Internal, "")
	}

	cur := row.Trip.Current.Location
	if req.Current != nil {
		cur = req.Current
	}

	row.Trip.Current = s.calcCurrentStatus(ctx, row.Trip.Current, cur)

	if req.EndTrip {
		row.Trip.End = row.Trip.Current
		row.Trip.Status = rentalpb.TripStatus_FINISHED
	}

	s.Mongo.UpdateTrip(ctx, tripId, accountId, row.UpdatedAt, row.Trip)
	return row.Trip, nil
}

func (s *Service) calcCurrentStatus(c context.Context, last *rentalpb.LocationStatus, current *rentalpb.Location) *rentalpb.LocationStatus {
	now := nowFunc()
	elapsedSec := float64(now - last.TimestampSec)

	dis, err := s.DistanceCalc.DistanceKm(c, last.Location, current)
	if err != nil {
		s.Logger.Warn("cannot calculate distance", zap.Error(err))
	}

	pointName, err := s.PointManager.Resolve(c, current)
	if err != nil {
		s.Logger.Warn("cannot resolve point name", zap.Stringer("location", current), zap.Error(err))
	}

	return &rentalpb.LocationStatus{
		Location:     current,
		FeeCent:      last.FeeCent + int32(centsPerSec*elapsedSec*2*rand.Float64()),
		KmDriven:     last.KmDriven + dis,
		TimestampSec: now,
		PointName:    pointName,
	}
}
