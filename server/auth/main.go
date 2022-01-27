package main

import (
	"coolcar/auth/api/gen/v1"
	"coolcar/auth/dao"
	"coolcar/auth/service"
	"coolcar/auth/wechat"
	"coolcar/shared/server"
	"crypto/rsa"
	"log"
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
	logger, err := server.NewZapLogger()
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

	ctx := context.Background()
	mgoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017/coolcar"))
	if err != nil {
		logger.Fatal("cannot connect to database", zap.Error(err))
	}

	logger.Sugar().Fatal(
		server.RunGRPCServer(&server.GRPCConifg{
			Name:   "auth",
			Addr:   ":8082",
			Logger: logger,
			RegisterFunc: func(s *grpc.Server) {
				authpb.RegisterAuthServiceServer(
					s,
					&service.OpenId{
						Logger: logger,
						Mongo:  dao.Use(mgoClient.Database("coolcar")),
						OpenIdProvider: &wechat.Remote{
							AppId:  appId,
							Secret: secret,
						},
						TokenGenerator: service.CreateTokenProvider(
							"coolcar/auth",
							&service.FilePrivateKeyProvider{},
						),
						TokenExpiresIn: 2 * time.Hour,
					},
				)
			},
		}),
	)
}
