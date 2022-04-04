package blob

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	"coolcar/blob/blob/dao"
	"coolcar/shared/id"
	"io"
	"io/ioutil"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Storage interface {
	SignUrl(ctx context.Context, httpMethod oss.HTTPMethod, path string, timeout time.Duration) (string, error)
	Get(ctx context.Context, path string) (io.ReadCloser, error)
}

type Service struct {
	OssId      string
	OssSecrets string
	Mongo      *dao.Mongo
	Logger     *zap.Logger
	Storage    Storage
	blobpb.UnimplementedBlobServiceServer
}

func (s *Service) getBlobRecord(c context.Context, blobId id.BlobId) (*dao.BLobRecord, error) {
	br, err := s.Mongo.GetBlob(c, blobId)
	if err == mongo.ErrNilDocument {
		return nil, status.Error(codes.NotFound, "")
	}

	if err != nil {
		s.Logger.Error("cannot get blob", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	return br, nil
}

func secToDuration(sec int32) time.Duration {
	return time.Duration(sec) * time.Second
}

func (s *Service) CreateBlob(ctx context.Context, req *blobpb.CreateBlobRequest) (*blobpb.CreateBlobResponse, error) {
	accountId := id.AccountId(req.AccountId)
	br, err := s.Mongo.CreateBlob(ctx, accountId)
	if err != nil {
		s.Logger.Error("cannot create blob: %v", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	url, err := s.Storage.SignUrl(ctx, oss.HTTPPut, br.Path, secToDuration(req.UploadUrlTimeoutSec))
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "cannot sign url: %v", err)
	}

	return &blobpb.CreateBlobResponse{
		Id:        br.Id.Hex(),
		UploadUrl: url,
	}, nil
}
func (s *Service) GetBlob(ctx context.Context, req *blobpb.GetBlobRequest) (*blobpb.GetBlobResponse, error) {
	br, err := s.getBlobRecord(ctx, id.BlobId(req.Id))
	if err != nil {
		return nil, err
	}

	r, err := s.Storage.Get(ctx, br.Path)
	if r != nil {
		defer r.Close()
	}
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "cannot get blob: %v", err)
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "cannot read from storage", err)
	}

	return &blobpb.GetBlobResponse{
		Data: b,
	}, nil
}

func (s *Service) GetBlobURL(ctx context.Context, req *blobpb.GetBlobURLRequest) (*blobpb.GetBlobURLResponse, error) {
	br, err := s.getBlobRecord(ctx, id.BlobId(req.Id))
	if err != nil {
		return nil, err
	}

	url, err := s.Storage.SignUrl(ctx, oss.HTTPGet, br.Path, secToDuration(req.TimeoutSec))
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "cannot sign URL: %v", err)
	}

	return &blobpb.GetBlobURLResponse{
		Url: url,
	}, nil
}
