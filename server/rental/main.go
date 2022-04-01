package main

import (
	"context"
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
	"os"

	"coolcar/rental/profile"
	profileDao "coolcar/rental/profile/dao"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	conn, err := grpc.Dial("localhost:18001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect aiservice", zap.Error(err))
	}

	logger.Sugar().Fatal(
		server.RunGRPCServer(&server.GRPCConifg{
			Name:              "rental",
			Addr:              ":8083",
			AuthPublicKeyPath: "shared/auth/public.key",
			Logger:            logger,
			RegisterFunc: func(s *grpc.Server) {
				db := mgoClient.Database("coolcar")
				rentalpb.RegisterTripServiceServer(
					s,
					&trip.Service{
						Logger:         logger,
						ProfileManager: &profileCli.Manager{},
						CarManager:     &car.Manager{},
						PointManager:   &poi.Manager{},
						Mongo:          tripDao.Use(db),
						DistanceCalc: &ai.Client{
							AIClient: coolenvpb.NewAIServiceClient(conn),
						},
					})

				rentalpb.RegisterProfileServiceServer(
					s,
					&profile.Service{
						Logger: logger,
						Mongo:  profileDao.Use(db),
					},
				)
			},
		}),
	)
}
