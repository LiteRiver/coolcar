package profile

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/rental/profile/dao"
	"coolcar/shared/auth"
	"coolcar/shared/id"
	mongotesting "coolcar/shared/mongo/testing"
	"coolcar/shared/server"
	"fmt"
	"os"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestProfileLifecycle(t *testing.T) {
	ctx := context.Background()
	s := newService(ctx, t)

	accountId := id.AccountId("account1")
	c := auth.ContextWithAccountId(ctx, accountId)
	cases := []struct {
		name       string
		op         func() (*rentalpb.Profile, error)
		wantName   string
		wantStatus rentalpb.IdentityStatus
		wantErr    bool
	}{
		{
			name: "get_empty",
			op: func() (*rentalpb.Profile, error) {
				return s.GetProfile(c, &rentalpb.GetProfileRequest{})
			},
			wantName:   "",
			wantStatus: rentalpb.IdentityStatus_UNSUBMITTED,
		},
		{
			name: "submit",
			op: func() (*rentalpb.Profile, error) {
				return s.SubmitProfile(c, &rentalpb.Identity{Name: "abc"})
			},
			wantName:   "abc",
			wantStatus: rentalpb.IdentityStatus_PENDING,
		},
		{
			name: "submit_again",
			op: func() (*rentalpb.Profile, error) {
				return s.SubmitProfile(c, &rentalpb.Identity{Name: "abc"})
			},
			wantName: "abc",
			wantErr:  true,
		},
		{
			name: "todo_force_verify",
			op: func() (*rentalpb.Profile, error) {
				p := &rentalpb.Profile{
					Identity: &rentalpb.Identity{
						Name: "abc",
					},
					IdentityStatus: rentalpb.IdentityStatus_VERIFIED,
				}

				err := s.Mongo.UpdateProfile(c, accountId, rentalpb.IdentityStatus_PENDING, p)
				if err != nil {
					return nil, err
				}
				return p, nil
			},
			wantName:   "abc",
			wantStatus: rentalpb.IdentityStatus_VERIFIED,
		},
		{
			name: "clear",
			op: func() (*rentalpb.Profile, error) {
				return s.ClearProfile(c, &rentalpb.ClearProfileRequest{})
			},
			wantName:   "",
			wantStatus: rentalpb.IdentityStatus_UNSUBMITTED,
		},
	}

	for _, cc := range cases {
		fmt.Printf("%s: begin testing\n", cc.name)
		p, err := cc.op()
		if cc.wantErr {
			if err == nil {
				t.Errorf("%s: wnat error; got none", cc.name)
			} else {
				continue
			}
		}

		if err != nil {
			t.Errorf("%s: operation failed: %v", cc.name, err)
			continue
		}

		if cc.wantName != "" && p.Identity.Name != cc.wantName {
			t.Errorf("%s: want name: %s, got name: %s", cc.name, cc.wantName, p.Identity.Name)
			continue
		}

		if p.IdentityStatus != cc.wantStatus {
			t.Errorf("%s: want status: %d, got status: %d", cc.name, cc.wantStatus, p.IdentityStatus)
			continue
		}
	}
}

func TestProfilePhotoLifecycle(t *testing.T) {
	ctx := auth.ContextWithAccountId(context.Background(), id.AccountId("account1"))
	s := newService(ctx, t)
	s.BlobClient = &blobClient{
		idForCreate: "blob1",
	}

	getPhotoOp := func() (string, error) {
		r, err := s.GetProfilePhoto(ctx, &rentalpb.GetProfilePhotoRequest{})
		if err != nil {
			return "", err
		}

		return r.Url, nil
	}

	cases := []struct {
		name        string
		op          func() (string, error)
		wantURL     string
		wantErrCode codes.Code
	}{
		{
			name:        "get_photo_before_upload",
			op:          getPhotoOp,
			wantErrCode: codes.NotFound,
		},
		{
			name: "create_blob",
			op: func() (string, error) {
				r, err := s.CreateProfilePhoto(ctx, &rentalpb.CreateProfilePhotoRequest{})
				if err != nil {
					return "", err
				}

				return r.UploadUrl, nil
			},
			wantURL: "upload_url for blob1",
		},
		{
			name: "complete_photo_upload",
			op: func() (string, error) {
				_, err := s.CompleteProfilePhoto(ctx, &rentalpb.CompleteProfilePhotoRequest{})
				return "", err
			},
			wantURL: "",
		},
		{
			name:    "get_phone_url",
			op:      getPhotoOp,
			wantURL: "get_url for blob1",
		},
		{
			name: "clear_photo",
			op: func() (string, error) {
				_, err := s.ClearProfilePhoto(ctx, &rentalpb.ClearProfilePhotoRequest{})
				return "", err
			},
		},
		{
			name:        "get_photo_after_clear",
			op:          getPhotoOp,
			wantErrCode: codes.NotFound,
		},
	}

	for _, cc := range cases {
		got, err := cc.op()
		code := codes.OK

		if err != nil {
			if s, ok := status.FromError(err); ok {
				code = s.Code()
			} else {
				t.Errorf("%s: operation failed: %v", cc.name, err)
			}
		}

		if code != cc.wantErrCode {
			t.Errorf("%s: wrong error code: want %d, got %d", cc.name, cc.wantErrCode, code)
		}

		if got != cc.wantURL {
			t.Errorf("%s: wrong url: want: %q, got: %q", cc.name, cc.wantURL, got)
		}
	}
}

func newService(ctx context.Context, t *testing.T) *Service {
	mc, err := mongotesting.NewClient(ctx)
	if err != nil {
		t.Fatalf("cannot connect to database: %v", err)
	}

	logger, err := server.NewZapLogger()
	if err != nil {
		t.Fatalf("cannot create logger: %v", err)
	}
	db := mc.Database("coolcar")
	mongotesting.SetupIndices(ctx, db)

	return &Service{
		Logger: logger,
		Mongo:  dao.Use(db),
	}
}

type blobClient struct {
	idForCreate string
}

func (b *blobClient) CreateBlob(ctx context.Context, in *blobpb.CreateBlobRequest, opts ...grpc.CallOption) (*blobpb.CreateBlobResponse, error) {
	return &blobpb.CreateBlobResponse{
		Id:        b.idForCreate,
		UploadUrl: "upload_url for " + b.idForCreate,
	}, nil
}

func (b *blobClient) GetBlob(ctx context.Context, in *blobpb.GetBlobRequest, opts ...grpc.CallOption) (*blobpb.GetBlobResponse, error) {
	return &blobpb.GetBlobResponse{}, nil
}

func (b *blobClient) GetBlobURL(ctx context.Context, in *blobpb.GetBlobURLRequest, opts ...grpc.CallOption) (*blobpb.GetBlobURLResponse, error) {
	return &blobpb.GetBlobURLResponse{
		Url: "get_url for " + in.Id,
	}, nil
}

func TestMain(m *testing.M) {
	os.Exit(mongotesting.RunWithMongoInDocker(m))
}
