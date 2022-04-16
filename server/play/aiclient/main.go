package main

import (
	"context"
	"coolcar/car/mq/amqpcli"
	coolenvpb "coolcar/shared/coolenv"
	"coolcar/shared/server"
	"encoding/json"
	"fmt"
	"time"

	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:18001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	ac := coolenvpb.NewAIServiceClient(conn)
	c := context.Background()
	res, err := ac.MeasureDistance(c, &coolenvpb.MeasureDistanceRequest{
		From: &coolenvpb.Location{
			Latitude:  30,
			Longitude: 120,
		},
		To: &coolenvpb.Location{
			Latitude:  31,
			Longitude: 121,
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", res)

	_, err = ac.SimulateCarPos(c, &coolenvpb.SimulateCarPosRequest{
		CarId: "car01",
		InitialPos: &coolenvpb.Location{
			Latitude:  30,
			Longitude: 120,
		},
		Type: coolenvpb.PosType_RANDOM,
	})

	if err != nil {
		panic(err)
	}

	logger, err := server.NewZapLogger()
	if err != nil {
		panic(err)
	}

	amqpConn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	if err != nil {
		panic(err)
	}

	sub, err := amqpcli.NewSubscriber(amqpConn, "pos_sim", logger)
	if err != nil {
		panic(err)
	}

	ch, cleanUp, err := sub.SubscribeRaw(c)
	defer cleanUp()
	if err != nil {
		panic(err)
	}

	tm := time.After(10 * time.Second)
	for {
		shouldStop := false
		select {
		case msg := <-ch:
			var update coolenvpb.CarPosUpdate
			err := json.Unmarshal(msg.Body, &update)
			if err != nil {
				fmt.Printf("cannot unmarshal car position: %v", err)
			} else {
				fmt.Printf("%+v\n", &update)
			}
		case <-tm:
			shouldStop = true
		}

		if shouldStop {
			break
		}
	}

	_, err = ac.EndSimulateCarPos(c, &coolenvpb.EndSimulateCarPosRequest{
		CarId: "car01",
	})
	if err != nil {
		panic(err)
	}
}
