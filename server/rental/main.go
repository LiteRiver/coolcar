package main

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	carpb "coolcar/car/api/gen/v1"
	"coolcar/rental/ai"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/rental/trip"
	"coolcar/rental/trip/client/car"
	"coolcar/rental/trip/client/poi"
	profileCli "coolcar/rental/trip/client/profile"
	tripDao "coolcar/rental/trip/dao"
	coolenvpb "coolcar/shared/coolenv"
	"coolcar/shared/server"
	"log"
	"time"

	"coolcar/rental/profile"
	profileDao "coolcar/rental/profile/dao"

	"github.com/namsral/flag"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var addr = flag.String("addr", ":8083", "address to listen")
var mongoURI = flag.String("mongo_uri", "mongodb://localhost:27017/coolcar", "mongo uri")
var aiAddr = flag.String("ai_addr", "localhost:18001", "address of ai service")
var carAddr = flag.String("car_addr", "localhost:8085", "address of car service")
var authPublicKeyFile = flag.String("auth_public_key_file", "shared/auth/public.key", "public key file path")

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

	conn, err := grpc.Dial(*aiAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect aiservice", zap.Error(err))
	}

	carConn, err := grpc.Dial(*carAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect car service", zap.Error(err))
	}

	logger.Sugar().Fatal(
		server.RunGRPCServer(&server.GRPCConifg{
			Name:              "rental",
			Addr:              *addr,
			AuthPublicKeyPath: *authPublicKeyFile,
			Logger:            logger,
			RegisterFunc: func(s *grpc.Server) {
				db := mgoClient.Database("coolcar")
				blobConn, err := grpc.Dial("localhost:8084", grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					logger.Fatal("cannot connect blob service", zap.Error(err))
				}

				profileService := &profile.Service{
					Logger:          logger,
					Mongo:           profileDao.Use(db),
					BlobClient:      blobpb.NewBlobServiceClient(blobConn),
					PhotoGetExpires: 5 * time.Second,
					PhotoPutExpires: 10 * time.Second,
				}

				rentalpb.RegisterTripServiceServer(
					s,
					&trip.Service{
						Logger: logger,
						ProfileManager: &profileCli.Manager{
							Provider: profileService,
						},
						CarManager: &car.Manager{
							CarService: carpb.NewCarServiceClient(carConn),
						},
						PointManager: &poi.Manager{},
						Mongo:        tripDao.Use(db),
						DistanceCalc: &ai.Client{
							AIClient: coolenvpb.NewAIServiceClient(conn),
						},
					},
				)

				rentalpb.RegisterProfileServiceServer(
					s,
					profileService,
				)
			},
		}),
	)
}
