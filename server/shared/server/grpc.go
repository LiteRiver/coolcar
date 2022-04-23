package server

import (
	"coolcar/shared/auth"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type GRPCConifg struct {
	Name              string
	Addr              string
	AuthPublicKeyPath string
	Logger            *zap.Logger
	RegisterFunc      func(*grpc.Server)
}

func RunGRPCServer(cfg *GRPCConifg) error {
	nameField := zap.String("name", cfg.Name)
	lis, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		cfg.Logger.Fatal("cannot listen", nameField, zap.Error(err))
	}

	opts := []grpc.ServerOption{}
	if cfg.AuthPublicKeyPath != "" {
		in, err := auth.Interceptor(cfg.AuthPublicKeyPath)
		if err != nil {
			cfg.Logger.Fatal("cannot create auth intercepter", zap.Error(err))
		}
		opts = append(opts, grpc.UnaryInterceptor(in))
	}

	svr := grpc.NewServer(opts...)

	grpc_health_v1.RegisterHealthServer(svr, health.NewServer())
	cfg.RegisterFunc(svr)
	cfg.Logger.Info("server started", nameField, zap.String("addr", cfg.Addr))
	return svr.Serve(lis)
}
