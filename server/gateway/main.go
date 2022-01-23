package main

import (
	"context"
	"coolcar/auth/api/gen/v1"
	rentalpb "coolcar/rental/api/gen/v1"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
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

	err := authpb.RegisterAuthServiceHandlerFromEndpoint(
		ctx,
		mux,
		"localhost:8082",
		[]grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	)

	if err != nil {
		log.Fatalf("cannot register auth service: %v\n", err)
	}

	err = rentalpb.RegisterTripServiceHandlerFromEndpoint(
		ctx,
		mux,
		"localhost:8083",
		[]grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	)

	if err != nil {
		log.Fatalf("cannot register trip service: %v\n", err)
	}
	
	log.Fatal(http.ListenAndServe(":8081", mux))
}
