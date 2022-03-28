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
	statusField    = tripField + ".status"
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

func (m *Mongo) CreateTrip(c context.Context, trip *rentalpb.Trip) (*TripRow, error) {
	r := &TripRow{
		Trip: trip,
	}

	r.Id = mgutil.NewObjectId()
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

func (m *Mongo) GetTrips(c context.Context, accountId id.AccountId, status rentalpb.TripStatus) ([]*TripRow, error) {
	filter := bson.M{
		accountIdField: accountId.String(),
	}

	if status != rentalpb.TripStatus_TS_NOT_SPECIFIED {
		filter[statusField] = status
	}

	res, err := m.col.Find(c, filter)
	if err != nil {
		return nil, err
	}

	var trips []*TripRow
	for res.Next(c) {
		var row TripRow
		err := res.Decode(&row)
		if err != nil {
			return nil, err
		}

		trips = append(trips, &row)
	}

	return trips, nil
}

func (m *Mongo) UpdateTrip(c context.Context, tripId id.TripId, accountId id.AccountId, updatedAt int64, trip *rentalpb.Trip) error {
	objId, err := objid.FromId(tripId)
	if err != nil {
		return fmt.Errorf("invalid id: %v", err)
	}

	newUpdatedAt := mgutil.UpdatedAt()
	res, err := m.col.UpdateOne(c, bson.M{
		mgutil.IdFieldName:        objId,
		accountIdField:            accountId.String(),
		mgutil.UpdatedAtFieldName: updatedAt,
	}, bson.M{
		"$set": bson.M{
			tripField:                 trip,
			mgutil.UpdatedAtFieldName: newUpdatedAt,
		},
	})

	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
