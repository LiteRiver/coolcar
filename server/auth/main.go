package main

import (
	"coolcar/auth/api/gen/v1"
	"coolcar/auth/dao"
	"coolcar/auth/service"
	"coolcar/auth/wechat"
	"crypto/rsa"
	"log"
	"net"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"golang.org/x/net/context"
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

	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		logger.Fatal("cannot listen", zap.Error(err))
	}

	ctx := context.Background()
	mgoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017/coolcar"))
	if err != nil {
		logger.Fatal("cannot connect to database", zap.Error(err))
	}

	svr := grpc.NewServer()
	var keyProvider PrivateKeyProvider = &service.FilePrivateKeyProvider{
		Logger: logger,
	}
	authpb.RegisterAuthServiceServer(
		svr, &service.OpenId{
			Logger: logger,
			Mongo:  dao.Use(mgoClient.Database("coolcar")),
			OpenIdProvider: &wechat.Remote{
				AppId:  appId,
				Secret: secret,
			},
			TokenGenerator: service.CreateTokenProvider("coolcar/auth", keyProvider.GetPrivateKey(logger)),
			TokenExpiresIn: 2 * time.Hour,
		},
	)

	err = svr.Serve(lis)
	logger.Fatal("cannot serve", zap.Error(err))
}
