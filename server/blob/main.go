package main

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	"coolcar/blob/blob"
	"coolcar/blob/blob/dao"
	"coolcar/blob/oss"
	"coolcar/shared/server"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger, err := server.NewZapLogger()
	if err != nil {
		log.Fatalf("cannot create logger: %v\n", err)
	}

	err = godotenv.Load()
	if err != nil {
		log.Fatalf("cannot load enviornment variables: %v\n", err)
	}

	ossAddr := os.Getenv("OSS_ADDR")
	if len(ossAddr) == 0 {
		log.Fatal("OSS_ADDR is empty")
	}

	ossId := os.Getenv("OSS_ID")
	if len(ossId) == 0 {
		log.Fatal("OSS_ID is empty")
	}

	ossSecrets := os.Getenv("OSS_SECRETS")
	if len(ossSecrets) == 0 {
		log.Fatal("OSS_SECRETS is empty")
	}

	ctx := context.Background()
	mgoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017/coolcar"))
	if err != nil {
		logger.Fatal("cannot connect to database", zap.Error(err))
	}

	st, err := oss.NewService(ossAddr, ossId, ossSecrets)
	if err != nil {
		logger.Fatal("cannot create OSS client", zap.Error(err))
	}

	logger.Sugar().Fatal(
		server.RunGRPCServer(&server.GRPCConifg{
			Name:   "blob",
			Addr:   ":8084",
			Logger: logger,
			RegisterFunc: func(s *grpc.Server) {
				db := mgoClient.Database("coolcar")
				blobpb.RegisterBlobServiceServer(
					s,
					&blob.Service{
						OssId:      ossId,
						OssSecrets: ossSecrets,
						Mongo:      dao.Use(db),
						Storage:    st,
						Logger:     logger,
					},
				)
			},
		}),
	)
}
