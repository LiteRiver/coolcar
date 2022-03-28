package dao

import (
	"context"
	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"coolcar/shared/mongo/objid"
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

func (mgo *Mongo) GetAccountId(ctx context.Context, openId string) (id.AccountId, error) {
	insertedId := mgutil.NewObjectId()
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

	return objid.ToAccountId(row.Id), nil
}
