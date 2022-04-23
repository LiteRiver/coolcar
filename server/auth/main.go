package main

import (
	"coolcar/auth/api/gen/v1"
	"coolcar/auth/dao"
	"coolcar/auth/service"
	"coolcar/auth/wechat"
	"coolcar/shared/server"
	"crypto/rsa"
	"fmt"
	"log"
	"time"

	"github.com/namsral/flag"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var addr = flag.String("addr", ":8082", "address to listen")
var mongoURI = flag.String("mongo_uri", "mongodb://localhost:27017/coolcar", "mongo URI")
var privateKeyFile = flag.String("private_key_file", "auth/private.key", "private key file path")
var wechatAppId = flag.String("wechat_app_id", "<wechat_app_id>", "wechat app id")
var wechatSecret = flag.String("wechat_secret", "<wechat_secret>", "wechat secret")

type PrivateKeyProvider interface {
	GetPrivateKey(logger *zap.Logger) *rsa.PrivateKey
}

func main() {
	flag.Parse()

	fmt.Println("mongo_uri:", *mongoURI)

	logger, err := server.NewZapLogger()
	if err != nil {
		log.Fatalf("cannot create logger: %v\n", err)
	}

	ctx := context.Background()
	mgoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(*mongoURI))
	if err != nil {
		logger.Fatal("cannot connect to database", zap.Error(err))
	}

	logger.Sugar().Fatal(
		server.RunGRPCServer(&server.GRPCConifg{
			Name:   "auth",
			Addr:   *addr,
			Logger: logger,
			RegisterFunc: func(s *grpc.Server) {
				authpb.RegisterAuthServiceServer(
					s,
					&service.OpenId{
						Logger: logger,
						Mongo:  dao.Use(mgoClient.Database("coolcar")),
						OpenIdProvider: &wechat.Remote{
							AppId:  *wechatAppId,
							Secret: *wechatSecret,
						},
						TokenGenerator: service.CreateTokenProvider(
							"coolcar/auth",
							&service.FilePrivateKeyProvider{
								PrivateKeyFile: *privateKeyFile,
							},
						),
						TokenExpiresIn: 2 * time.Hour,
					},
				)
			},
		}),
	)
}
