package dao

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/id"
	"coolcar/shared/mongo/objid"
	mongotesting "coolcar/shared/mongo/testing"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoURI string

func TestCreateTrip(t *testing.T) {
	mongoURI = "mongodb://localhost:27017"
	ctx := context.Background()
	mc, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Fatalf("cannot connect to database: %v", err)
	}

	mgo := Use(mc.Database("coolcar"))
	acctId := id.AccountId("account1")
	row, err := mgo.CreateTrip(ctx, &rentalpb.Trip{
		AccountId: acctId.String(),
		CarId:     "car1",
		Start: &rentalpb.LocationStatus{
			PointName: "start1",
			Location: &rentalpb.Location{
				Latitude:  30,
				Longitude: 120,
			},
		},
		End: &rentalpb.LocationStatus{
			PointName: "endpoint",
			FeeCent:   10000,
			KmDriven:  35,
			Location: &rentalpb.Location{
				Latitude:  35,
				Longitude: 115,
			},
		},
		Status: rentalpb.TripStatus_FINISHED,
	})

	if err != nil {
		t.Errorf("can not create trip: %v", err)
	}

	t.Errorf("inserted row %s with updatedat %v", row.Id, row.UpdatedAt)

	got, err := mgo.GetTrip(ctx, objid.ToTripId(row.Id), acctId)
	if err != nil {
		t.Errorf("can not get trip: %v", err)
	}

	t.Errorf("got trip: %+v", got)
}

func TestMain(m *testing.M) {
	os.Exit(mongotesting.RunWithMongoInDocker(m, &mongoURI))
}
