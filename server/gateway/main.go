package main

import (
	"context"
	"coolcar/auth/api/gen/v1"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/server"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	logger, err := server.NewZapLogger()
	if err != nil {
		log.Fatalf("cannot create logger: %v", err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(
			runtime.MIMEWildcard,
			&runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{UseEnumNumbers: true, UseProtoNames: true},
			},
		),
	)

	serverConfigs := []struct {
		name         string
		addr         string
		registerFunc func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error
	}{
		{
			name:         "auth",
			addr:         "localhost:8082",
			registerFunc: authpb.RegisterAuthServiceHandlerFromEndpoint,
		},
		{
			name:         "rental",
			addr:         "localhost:8083",
			registerFunc: rentalpb.RegisterTripServiceHandlerFromEndpoint,
		},
	}

	for _, cfg := range serverConfigs {
		err := cfg.registerFunc(ctx, mux, cfg.addr, []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})

		if err != nil {
			logger.Sugar().Fatalf("cannot register service %s", cfg.name)
		}
	}

	addr := ":8081"
	logger.Sugar().Infof("grpc gateway started at: %s", addr)
	logger.Sugar().Fatal(http.ListenAndServe(addr, mux))
}
