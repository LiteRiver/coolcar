package main

import (
	"context"
	"coolcar/car/amqpcli"
	carpb "coolcar/car/api/gen/v1"
	"coolcar/car/car"
	"coolcar/car/car/dao"
	"coolcar/car/sim"
	"coolcar/shared/server"
	"log"

	"github.com/streadway/amqp"
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

	ctx := context.Background()
	mgoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017/coolcar"))
	if err != nil {
		logger.Fatal("cannot connect to database", zap.Error(err))
	}

	dmqpConn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	if err != nil {
		logger.Fatal("cannot dial dmqp", zap.Error(err))
	}

	exchange := "coolcar"
	pub, err := amqpcli.NewPublisher(dmqpConn, exchange)
	if err != nil {
		logger.Fatal("cannot create publisher", zap.Error(err))
	}

	carConn, err := grpc.Dial("localhost:8085", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect car service", zap.Error(err))
	}

	sub, err := amqpcli.NewSubscriber(dmqpConn, exchange, logger)
	if err != nil {
		logger.Fatal("cannot connect create subscriber", zap.Error(err))
	}

	simController := &sim.Controller{
		CarService: carpb.NewCarServiceClient(carConn),
		Logger:     logger,
		Subscriber: sub,
	}
	go simController.RunSimulations(context.Background())

	logger.Sugar().Fatal(
		server.RunGRPCServer(&server.GRPCConifg{
			Name:   "car",
			Addr:   ":8085",
			Logger: logger,
			RegisterFunc: func(s *grpc.Server) {
				db := mgoClient.Database("coolcar")

				carpb.RegisterCarServiceServer(
					s,
					&car.Service{
						Logger:    logger,
						Mongo:     dao.Use(db),
						Publisher: pub,
					},
				)
			},
		}),
	)
}
