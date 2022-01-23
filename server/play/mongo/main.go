package main

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

func main() {
	ctx := context.Background()
	mc, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017/coolcar"))
	if err != nil {
		panic(err)
	}
	col := mc.Database("coolcar").Collection("accounts")

	findRows(ctx, col)
}

func findRows(ctx context.Context, col *mongo.Collection) {
	cur, err := col.Find(ctx, bson.M{})
	if err != nil {
		panic(err)
	}

	for cur.Next(ctx) {
		var row struct {
			ID     primitive.ObjectID `bson:"_id"`
			OpenId string             `bson:"open_id"`
		}
		err = cur.Decode(&row)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%+v\n", row)
	}
	// ret := col.FindOne(ctx, bson.M{
	// 	"open_id": "123",
	// })

	// var row struct {
	// 	ID     primitive.ObjectID `bson:"_id"`
	// 	OpenId string             `bson:"open_id"`
	// }

	// err := ret.Decode(&row)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Printf("%+v\n", row)
}

func insertRows(ctx context.Context, col *mongo.Collection) {
	ret, err := col.InsertMany(ctx, []interface{}{
		bson.M{
			"open_id": "123",
		},
		bson.M{
			"open_id": "456",
		},
	})

	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", ret)

}
