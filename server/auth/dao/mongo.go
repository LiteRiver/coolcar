package dao

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	col         *mongo.Collection
	newObjectId func() primitive.ObjectID
}

func Use(db *mongo.Database) *Mongo {
	return &Mongo{
		col:         db.Collection("accounts"),
		newObjectId: primitive.NewObjectID,
	}
}

func (mgo *Mongo) GetAccountId(ctx context.Context, openId string) (string, error) {
	insertedId := mgo.newObjectId()
	ret := mgo.col.FindOneAndUpdate(ctx, bson.M{
		"open_id": openId,
	},
		bson.M{
			// "$set": bson.M{
			// 	"open_id": openId,
			// },
			"$setOnInsert": bson.M{
				"_id":     insertedId,
				"open_id": openId,
			},
		},
		options.FindOneAndUpdate().
			SetUpsert(true).
			SetReturnDocument(options.After),
	)

	if err := ret.Err(); err != nil {
		return "", fmt.Errorf("cannot findOneAndUpdate: %v", err)
	}

	var row struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	err := ret.Decode(&row)
	if err != nil {
		return "", fmt.Errorf("cannot decode result: %v", err)
	}

	return row.ID.Hex(), nil
}
