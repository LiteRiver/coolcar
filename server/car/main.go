package main

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	"coolcar/car/car"
	"coolcar/car/car/dao"
	"coolcar/car/mq/amqpcli"
	"coolcar/car/sim"
	"coolcar/car/sim/pos"
	"coolcar/car/trip"
	"coolcar/car/ws"
	rentalpb "coolcar/rental/api/gen/v1"
	coolenvpb "coolcar/shared/coolenv"
	"coolcar/shared/server"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
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

	amqpConn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	if err != nil {
		logger.Fatal("cannot dial dmqp", zap.Error(err))
	}

	exchange := "coolcar"
	pub, err := amqpcli.NewPublisher(amqpConn, exchange)
	if err != nil {
		logger.Fatal("cannot create publisher", zap.Error(err))
	}

	carConn, err := grpc.Dial("localhost:8085", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect car service", zap.Error(err))
	}

	aiConn, err := grpc.Dial("localhost:18001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect ai service", zap.Error(err))
	}

	carSub, err := amqpcli.NewSubscriber(amqpConn, exchange, logger)
	if err != nil {
		logger.Fatal("cannot connect create car subscriber", zap.Error(err))
	}

	posSub, err := amqpcli.NewSubscriber(amqpConn, "pos_sim", logger)
	if err != nil {
		logger.Fatal("cannot create position subscriber", zap.Error(err))
	}

	simController := &sim.Controller{
		CarService:    carpb.NewCarServiceClient(carConn),
		AIService:     coolenvpb.NewAIServiceClient(aiConn),
		Logger:        logger,
		CarSubscriber: carSub,
		PositionSubscriber: &pos.Subscriber{
			Sub:    posSub,
			Logger: logger,
		},
	}
	go simController.RunSimulations(context.Background())

	u := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	tripConn, err := grpc.Dial("localhost:8083", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect trip service", zap.Error(err))
	}

	go trip.RunUpdater(carSub, rentalpb.NewTripServiceClient(tripConn), logger)

	http.HandleFunc("/ws", ws.Handler(u, carSub, logger))
	go func() {
		addr := ":9000"
		logger.Info("HTTP server started.", zap.String("addr", addr))
		logger.Sugar().Fatal(http.ListenAndServe(addr, nil))
	}()

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
