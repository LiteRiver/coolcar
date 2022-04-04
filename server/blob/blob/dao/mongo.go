package dao

import (
	"context"
	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"coolcar/shared/mongo/objid"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Mongo struct {
	col         *mongo.Collection
	newObjectId func() primitive.ObjectID
}

func Use(db *mongo.Database) *Mongo {
	return &Mongo{
		col:         db.Collection("blobs"),
		newObjectId: primitive.NewObjectID,
	}
}

type BLobRecord struct {
	mgutil.IdField `bson:"inline"`
	AccountId      string `bson:"accountid"`
	Path           string `bson:"path"`
}

func (m *Mongo) CreateBlob(c context.Context, accountId id.AccountId) (*BLobRecord, error) {
	br := &BLobRecord{
		AccountId: accountId.String(),
	}
	id := mgutil.NewObjectId()
	br.Id = id
	br.Path = fmt.Sprintf("%s/%s", accountId.String(), id.Hex())
	_, err := m.col.InsertOne(c, br)

	if err != nil {
		return nil, err
	}

	return br, nil
}

func (m *Mongo) GetBlob(c context.Context, blobId id.BlobId) (*BLobRecord, error) {
	id, err := objid.FromId((blobId))
	if err != nil {
		return nil, fmt.Errorf("invalid object id: %v", err)
	}

	res := m.col.FindOne(
		c,
		bson.M{
			mgutil.IdFieldName: id,
		},
	)

	if err := res.Err(); err != nil {
		return nil, err
	}

	var br BLobRecord
	err = res.Decode(&br)
	if err != nil {
		return nil, fmt.Errorf("cannot decode blob record: %v", err)
	}

	return &br, nil
}
