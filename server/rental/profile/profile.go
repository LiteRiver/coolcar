package profile

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/rental/profile/dao"
	"coolcar/shared/auth"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	Logger *zap.Logger
	Mongo  *dao.Mongo
	rentalpb.UnimplementedProfileServiceServer
}

func (s *Service) GetProfile(ctx context.Context, req *rentalpb.GetProfileRequest) (*rentalpb.Profile, error) {
	accountId, err := auth.AccountIdFromContext(ctx)
	if err != nil {
		return nil, err
	}

	profile, err := s.Mongo.GetProfile(ctx, accountId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &rentalpb.Profile{}, nil
		}
		s.Logger.Error("cannot get profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	return profile, nil
}

// TODO: always receive empty Identity
func (s *Service) SubmitProfile(ctx context.Context, identity *rentalpb.Identity) (*rentalpb.Profile, error) {
	accountId, err := auth.AccountIdFromContext(ctx)
	if err != nil {
		return nil, err
	}
	profile := &rentalpb.Profile{
		Identity:       identity,
		IdentityStatus: rentalpb.IdentityStatus_PENDING,
	}
	err = s.Mongo.UpdateProfile(ctx, accountId, rentalpb.IdentityStatus_UNSUBMITTED, profile)
	if err != nil {
		s.Logger.Error("cannot update profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}
	go func() {
		time.Sleep(3 * time.Second)
		err := s.Mongo.UpdateProfile(context.Background(), accountId, rentalpb.IdentityStatus_PENDING, &rentalpb.Profile{
			Identity:       identity,
			IdentityStatus: rentalpb.IdentityStatus_VERIFIED,
		})
		if err != nil {
			s.Logger.Error("cannot verify identity: %v", zap.Error(err))
		}
	}()
	return profile, nil
}
func (s *Service) ClearProfile(ctx context.Context, req *rentalpb.ClearProfileRequest) (*rentalpb.Profile, error) {
	accountId, err := auth.AccountIdFromContext(ctx)
	if err != nil {
		return nil, err
	}
	profile := &rentalpb.Profile{}
	err = s.Mongo.UpdateProfile(ctx, accountId, rentalpb.IdentityStatus_VERIFIED, profile)
	if err != nil {
		s.Logger.Error("cannot update profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}
	return profile, nil
}
