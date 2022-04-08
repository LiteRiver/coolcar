package dao

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"coolcar/shared/mongo/objid"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	carField      = "car"
	statusField   = carField + ".status"
	driverField   = carField + ".driver"
	positionField = carField + ".position"
	tripIdField   = carField + ".tripid"
)

type Mongo struct {
	col         *mongo.Collection
	newObjectId func() primitive.ObjectID
}

func Use(db *mongo.Database) *Mongo {
	return &Mongo{
		col:         db.Collection("cars"),
		newObjectId: primitive.NewObjectID,
	}
}

type CarRow struct {
	mgutil.IdField `bson:"inline"`
	Car            *carpb.Car `bson:"car"`
}

func (m *Mongo) CreateCar(c context.Context) (*CarRow, error) {
	cr := &CarRow{
		Car: &carpb.Car{
			Status: carpb.CarStatus_LOCKED,
			Position: &carpb.Location{
				Latitude:  30,
				Longitude: 120,
			},
		},
	}

	cr.Id = mgutil.NewObjectId()
	_, err := m.col.InsertOne(c, cr)
	if err != nil {
		return nil, err
	}

	return cr, nil
}

func (m *Mongo) GetCar(c context.Context, id id.CarId) (*CarRow, error) {
	carId, err := objid.FromId(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %v", err)
	}
	ret := m.col.FindOne(c, bson.M{
		mgutil.IdFieldName: carId,
	})

	return convertSingleResult(ret)
}

func (m *Mongo) GetCars(c context.Context) ([]*CarRow, error) {
	filter := bson.M{}
	cur, err := m.col.Find(c, filter, options.Find())
	if err != nil {
		return nil, err
	}

	var crs []*CarRow
	for cur.Next(c) {
		var cr CarRow
		err := cur.Decode(&cr)
		if err != nil {
			return nil, err
		}

		crs = append(crs, &cr)
	}

	return crs, nil
}

type CarUpdate struct {
	Status       carpb.CarStatus
	Position     *carpb.Location
	Driver       *carpb.Driver
	UpdateTripId bool
	TripId       id.TripId
}

func (m *Mongo) UpdateCar(c context.Context, id id.CarId, prevStatus carpb.CarStatus, update *CarUpdate) (*CarRow, error) {
	objId, err := objid.FromId(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %v", err)
	}

	filter := bson.M{
		mgutil.IdFieldName: objId,
	}

	if prevStatus != carpb.CarStatus_CS_NOT_SPECIFIED {
		filter[statusField] = prevStatus
	}

	u := bson.M{}
	if update.Status != carpb.CarStatus_CS_NOT_SPECIFIED {
		u[statusField] = update.Status
	}

	if update.Driver != nil {
		u[driverField] = update.Driver
	}

	if update.Position != nil {
		u[positionField] = update.Position
	}

	if update.UpdateTripId {
		u[tripIdField] = update.TripId
	}

	ret := m.col.FindOneAndUpdate(
		c,
		filter,
		mgutil.Set(u),
		options.
			FindOneAndUpdate().
			SetReturnDocument(options.After),
	)

	return convertSingleResult(ret)
}

func convertSingleResult(ret *mongo.SingleResult) (*CarRow, error) {
	if err := ret.Err(); err != nil {
		return nil, err
	}

	var cr CarRow
	err := ret.Decode(&cr)
	if err != nil {
		return nil, fmt.Errorf("cannot decode car row: %v", err)
	}

	return &cr, nil
}
