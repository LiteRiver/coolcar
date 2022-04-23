package main

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	"coolcar/blob/blob"
	"coolcar/blob/blob/dao"
	"coolcar/blob/oss"
	"coolcar/shared/server"
	"log"

	"github.com/namsral/flag"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var addr = flag.String("addr", ":8084", "address to listen")
var mongoURI = flag.String("mongo_uri", "mongodb://localhost:27017/coolcar", "mongo URI")
var ossAddr = flag.String("oss_addr", "<oss_addr>", "address of OSS")
var ossId = flag.String("oss_id", "<oss_id>", "id of OSS")
var ossSecrets = flag.String("oss_secrets", "<oss_secrets>", "secrets of OSS")

func main() {
	flag.Parse()

	logger, err := server.NewZapLogger()
	if err != nil {
		log.Fatalf("cannot create logger: %v\n", err)
	}

	ctx := context.Background()
	mgoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(*mongoURI))
	if err != nil {
		logger.Fatal("cannot connect to database", zap.Error(err))
	}

	st, err := oss.NewService(*ossAddr, *ossId, *ossSecrets)
	if err != nil {
		logger.Fatal("cannot create OSS client", zap.Error(err))
	}

	logger.Sugar().Fatal(
		server.RunGRPCServer(&server.GRPCConifg{
			Name:   "blob",
			Addr:   *addr,
			Logger: logger,
			RegisterFunc: func(s *grpc.Server) {
				db := mgoClient.Database("coolcar")
				blobpb.RegisterBlobServiceServer(
					s,
					&blob.Service{
						OssId:      *ossId,
						OssSecrets: *ossSecrets,
						Mongo:      dao.Use(db),
						Storage:    st,
						Logger:     logger,
					},
				)
			},
		}),
	)
}
