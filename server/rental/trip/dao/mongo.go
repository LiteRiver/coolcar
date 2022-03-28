package dao

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"coolcar/shared/mongo/objid"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	tripField      = "trip"
	accountIdField = tripField + ".accountid"
)

type Mongo struct {
	col         *mongo.Collection
	newObjectId func() primitive.ObjectID
}

func Use(db *mongo.Database) *Mongo {
	return &Mongo{
		col:         db.Collection("trips"),
		newObjectId: primitive.NewObjectID,
	}
}

type TripRow struct {
	mgutil.IdField        `bson:"inline"`
	mgutil.UpdatedAtField `bson:"inline"`
	Trip                  *rentalpb.Trip `bson:"trip"`
}

// TODO: 同一个account最多只能有个一个进行中的Trip
// TODO: 强类型化TripId
// TODO: 表格驱动测试

func (m *Mongo) CreateTrip(c context.Context, trip *rentalpb.Trip) (*TripRow, error) {
	r := &TripRow{
		Trip: trip,
	}

	r.Id = mgutil.NewObjectID()
	r.UpdatedAt = mgutil.UpdatedAt()
	_, err := m.col.InsertOne(c, r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (m *Mongo) GetTrip(c context.Context, id id.TripId, accountId id.AccountId) (*TripRow, error) {
	objId, err := objid.FromId(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %v", err)
	}
	res := m.col.FindOne(c, bson.M{
		mgutil.IdFieldName: objId,
		accountIdField:     accountId,
	})

	if err = res.Err(); err != nil {
		return nil, err
	}

	var row TripRow
	err = res.Decode(&row)
	if err != nil {
		return nil, fmt.Errorf("can not decode: %v", err)
	}

	return &row, nil
}
