package dao

import (
	"context"
	mgutil "coolcar/shared/mongo"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	col *mongo.Collection
}

func Use(db *mongo.Database) *Mongo {
	return &Mongo{
		col: db.Collection("accounts"),
	}
}

func (mgo *Mongo) GetAccountId(ctx context.Context, openId string) (string, error) {
	insertedId := mgutil.NewObjectID()
	ret := mgo.col.FindOneAndUpdate(ctx, bson.M{
		"open_id": openId,
	},
		bson.M{
			// "$set": bson.M{
			// 	"open_id": openId,
			// },
			"$setOnInsert": bson.M{
				mgutil.IdFieldName: insertedId,
				"open_id":          openId,
			},
		},
		options.FindOneAndUpdate().
			SetUpsert(true).
			SetReturnDocument(options.After),
	)

	if err := ret.Err(); err != nil {
		return "", fmt.Errorf("cannot findOneAndUpdate: %v", err)
	}

	var row mgutil.IdField
	err := ret.Decode(&row)
	if err != nil {
		return "", fmt.Errorf("cannot decode result: %v", err)
	}

	return row.Id.Hex(), nil
}
