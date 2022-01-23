package main

import (
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/rental/trip"
	"crypto/rsa"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type PrivateKeyProvider interface {
	GetPrivateKey(logger *zap.Logger) *rsa.PrivateKey
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("cannot create logger: %v\n", err)
	}
	err = godotenv.Load()
	if err != nil {
		log.Fatalf("cannot load enviornment variables: %v\n", err)
	}

	appId := os.Getenv("APP_ID")
	if len(appId) == 0 {
		log.Fatal("APP_ID is empty")
	}
	secret := os.Getenv("SECRET")
	if len(secret) == 0 {
		log.Fatal("SECRET is empty")
	}

	lis, err := net.Listen("tcp", ":8083")
	if err != nil {
		logger.Fatal("cannot listen", zap.Error(err))
	}

	svr := grpc.NewServer()
	rentalpb.RegisterTripServiceServer(
		svr, &trip.Service{
			Logger: logger,
		})

	err = svr.Serve(lis)
	logger.Fatal("cannot serve", zap.Error(err))
}
