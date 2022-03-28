package service

import (
	"context"
	"coolcar/auth/api/gen/v1"
	"coolcar/auth/dao"
	"coolcar/auth/wechat"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OpenId struct {
	OpenIdProvider
	TokenGenerator
	TokenExpiresIn time.Duration
	Logger         *zap.Logger
	Mongo          *dao.Mongo
	authpb.UnimplementedAuthServiceServer
}

type OpenIdProvider interface {
	GetOpenId(code string) (*wechat.OpenIdResponse, error)
}

type TokenGenerator interface {
	GenerateToken(accountId string, expiresIn time.Duration) (string, error)
}

func (svc *OpenId) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	res, err := svc.GetOpenId(req.Code)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "cannot get openid: %v", err)
	}

	accountId, err := svc.Mongo.GetAccountId(ctx, res.OpenId)
	if err != nil {
		svc.Logger.Error("cannot get account id", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	token, err := svc.TokenGenerator.GenerateToken(accountId.String(), svc.TokenExpiresIn)
	if err != nil {
		svc.Logger.Error("cannot generate token", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	return &authpb.LoginResponse{
		AccessToken: token,
		ExpiresIn:   int32(svc.TokenExpiresIn.Seconds()),
	}, nil
}
