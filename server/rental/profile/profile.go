package profile

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/rental/profile/dao"
	"coolcar/shared/auth"
	"coolcar/shared/id"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	Logger          *zap.Logger
	Mongo           *dao.Mongo
	BlobClient      blobpb.BlobServiceClient
	PhotoGetExpires time.Duration
	PhotoPutExpires time.Duration
	rentalpb.UnimplementedProfileServiceServer
}

func (s *Service) GetProfile(ctx context.Context, req *rentalpb.GetProfileRequest) (*rentalpb.Profile, error) {
	accountId, err := auth.AccountIdFromContext(ctx)
	if err != nil {
		return nil, err
	}

	pr, err := s.Mongo.GetProfile(ctx, accountId)
	if err != nil {
		code := s.logAndConvertProfileErr(err)
		if code == codes.NotFound {
			return &rentalpb.Profile{}, nil
		} else {
			return nil, status.Error(code, "")
		}
	}

	if pr.Profile == nil {
		return &rentalpb.Profile{}, nil
	}

	return pr.Profile, nil
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
func (s *Service) GetProfilePhoto(ctx context.Context, req *rentalpb.GetProfilePhotoRequest) (*rentalpb.GetProfilePhotoResponse, error) {
	accountId, err := auth.AccountIdFromContext(ctx)
	if err != nil {
		return nil, err
	}

	pr, err := s.Mongo.GetProfile(ctx, accountId)
	if err != nil {
		return nil, status.Error(s.logAndConvertProfileErr(err), "")
	}
	if pr.PhotoBlobId == "" {
		return nil, status.Error(codes.NotFound, "")
	}

	br, err := s.BlobClient.GetBlobURL(ctx, &blobpb.GetBlobURLRequest{Id: pr.PhotoBlobId, TimeoutSec: int32(s.PhotoGetExpires.Seconds())})
	if err != nil {
		s.Logger.Error("cannot get blob", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	return &rentalpb.GetProfilePhotoResponse{
		Url: br.Url,
	}, nil
}

func (s *Service) CreateProfilePhoto(ctx context.Context, req *rentalpb.CreateProfilePhotoRequest) (*rentalpb.CreateProfilePhotoResponse, error) {
	accountId, err := auth.AccountIdFromContext(ctx)
	if err != nil {
		return nil, err
	}

	br, err := s.BlobClient.CreateBlob(ctx, &blobpb.CreateBlobRequest{
		AccountId:           accountId.String(),
		UploadUrlTimeoutSec: int32(s.PhotoPutExpires.Seconds()),
	})

	if err != nil {
		s.Logger.Error("cannot create blob", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	err = s.Mongo.UpdateProfilePhoto(ctx, accountId, id.BlobId(br.Id))
	if err != nil {
		s.Logger.Error("cannot update profile photo", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	return &rentalpb.CreateProfilePhotoResponse{
		UploadUrl: br.UploadUrl,
	}, nil
}

func (s *Service) CompleteProfilePhoto(ctx context.Context, req *rentalpb.CompleteProfilePhotoRequest) (*rentalpb.Identity, error) {
	accountId, err := auth.AccountIdFromContext(ctx)
	if err != nil {
		return nil, err
	}

	pr, err := s.Mongo.GetProfile(ctx, accountId)
	if err != nil {
		return nil, status.Error(s.logAndConvertProfileErr(err), "")
	}
	if pr.PhotoBlobId == "" {
		return nil, status.Error(codes.NotFound, "")
	}

	br, err := s.BlobClient.GetBlob(ctx, &blobpb.GetBlobRequest{Id: pr.PhotoBlobId})
	if err != nil {
		s.Logger.Error("cannot get blob", zap.Error(err))
		return nil, status.Error(codes.Aborted, "")
	}

	s.Logger.Info("got profile photo", zap.Int("size", len(br.Data)))

	return &rentalpb.Identity{
		LicenseNumber: "333333333",
		Name:          "李四",
		Gender:        rentalpb.Gender_MALE,
		DateOfBirthMs: 1649130734000,
	}, nil
}

func (s *Service) ClearProfilePhoto(ctx context.Context, req *rentalpb.ClearProfilePhotoRequest) (*rentalpb.ClearProfilePhotoResponse, error) {
	accountId, err := auth.AccountIdFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = s.Mongo.UpdateProfilePhoto(ctx, accountId, id.BlobId(""))
	if err != nil {
		s.Logger.Error("cannot update profile photo", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	return &rentalpb.ClearProfilePhotoResponse{}, nil
}

func (s *Service) logAndConvertProfileErr(err error) codes.Code {
	if err == mongo.ErrNoDocuments {
		return codes.NotFound
	}
	s.Logger.Error("cannot get profile", zap.Error(err))
	return codes.Internal
}
