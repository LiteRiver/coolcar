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
	Logger *zap.Logger
	Mongo  *dao.Mongo
	rentalpb.UnimplementedTripServiceServer
}

func (s *Service) CreateTrip(ctx context.Context, req *rentalpb.CreateTripRequest) (*rentalpb.CreateTripResponse, error) {
	accountId, err := auth.AccountIdFromContext(ctx)
	if err != nil {
		return nil, err
	}

	s.Logger.Info("create trip", zap.Any("start", req.Start), zap.String("account_id:", accountId.String()))
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *Service) UpdateTrip(ctx context.Context, req *rentalpb.UpdateTripRequest) (*rentalpb.Trip, error) {
	accountId, err := auth.AccountIdFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "")
	}
	tripId := id.TripId(req.Id)
	row, err := s.Mongo.GetTrip(ctx, tripId, accountId)

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
