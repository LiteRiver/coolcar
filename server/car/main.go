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
	"github.com/namsral/flag"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var addr = flag.String("addr", ":8085", "address to listen")
var wsAddr = flag.String("ws_addr", ":9000", "websocket address to listen")
var mongoURI = flag.String("mongo_uri", "mongodb://localhost:27017/coolcar", "mongo URI")
var amqpURI = flag.String("amqp_uri", "amqp://guest:guest@localhost:5672", "amqp URI")
var carAddr = flag.String("car_addr", "localhost:8085", "address of car service")
var aiAddr = flag.String("ai_addr", "localhost:18001", "address of ai service")
var tripAddr = flag.String("trip_addr", "localhost:8083", "address of trip service")

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

	amqpConn, err := amqp.Dial(*amqpURI)
	if err != nil {
		logger.Fatal("cannot dial dmqp", zap.Error(err))
	}

	exchange := "coolcar"
	pub, err := amqpcli.NewPublisher(amqpConn, exchange)
	if err != nil {
		logger.Fatal("cannot create publisher", zap.Error(err))
	}

	carConn, err := grpc.Dial(*carAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect car service", zap.Error(err))
	}

	aiConn, err := grpc.Dial(*aiAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	tripConn, err := grpc.Dial(*tripAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect trip service", zap.Error(err))
	}

	go trip.RunUpdater(carSub, rentalpb.NewTripServiceClient(tripConn), logger)

	http.HandleFunc("/ws", ws.Handler(u, carSub, logger))
	go func() {
		addr := *wsAddr
		logger.Info("HTTP server started.", zap.String("addr", addr))
		logger.Sugar().Fatal(http.ListenAndServe(addr, nil))
	}()

	logger.Sugar().Fatal(
		server.RunGRPCServer(&server.GRPCConifg{
			Name:   "car",
			Addr:   *addr,
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
