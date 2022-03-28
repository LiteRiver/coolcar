package dao

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"coolcar/shared/mongo/objid"
	mongotesting "coolcar/shared/mongo/testing"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/testing/protocmp"
)

// var mnongoURI string

func TestCreateTrip(t *testing.T) {
	ctx := context.Background()
	mc, err := mongotesting.NewClient(ctx)
	if err != nil {
		t.Fatalf("cannot connect to database: %v", err)
	}

	db := mc.Database("coolcar")
	err = mongotesting.SetupIndices(ctx, db)

	if err != nil {
		t.Fatalf("cannot setup indices: %v", err)
	}

	cases := []struct {
		name       string
		tripId     string
		accountId  string
		tripStatus rentalpb.TripStatus
		wantErr    bool
	}{
		{
			name:       "finished",
			tripId:     "61f2b7df729f5d8b3bc69be4",
			accountId:  "account1",
			tripStatus: rentalpb.TripStatus_FINISHED,
			wantErr:    false,
		},
		{
			name:       "another_finished",
			tripId:     "61f2b7df729f5d8b3bc69be5",
			accountId:  "account1",
			tripStatus: rentalpb.TripStatus_FINISHED,
			wantErr:    false,
		},
		{
			name:       "in_progress",
			tripId:     "61f2b7df729f5d8b3bc69be6",
			accountId:  "account1",
			tripStatus: rentalpb.TripStatus_IN_PROGRESS,
			wantErr:    false,
		},
		{
			name:       "another_in_progress",
			tripId:     "61f2b7df729f5d8b3bc69be7",
			accountId:  "account1",
			tripStatus: rentalpb.TripStatus_IN_PROGRESS,
			wantErr:    true,
		},
		{
			name:       "in_progress_by_another_account",
			tripId:     "61f2b7df729f5d8b3bc69be8",
			accountId:  "account2",
			tripStatus: rentalpb.TripStatus_IN_PROGRESS,
			wantErr:    false,
		},
	}

	mgo := Use(db)
	for _, cc := range cases {
		mgutil.NewObjectID = func() primitive.ObjectID {
			return objid.EnsureObjId(id.TripId(cc.tripId))
		}

		row, err := mgo.CreateTrip(ctx, &rentalpb.Trip{
			AccountId: cc.accountId,
			Status:    cc.tripStatus,
		})

		if cc.wantErr {
			if err == nil {
				t.Errorf("%s: error expected; got none", cc.name)
			}
			continue
		}

		if err != nil {
			t.Errorf("%s: error creating trip: %v", cc.name, err)
			continue
		}

		if row.Id.Hex() != cc.tripId {
			t.Errorf("%s: incorrect trip id; want: %s, got: %s", cc.name, cc.tripId, row.Id.Hex())
		}
	}
}

func TestGetTrip(t *testing.T) {
	ctx := context.Background()
	mc, err := mongotesting.NewClient(ctx)
	if err != nil {
		t.Fatalf("cannot connect to database: %v", err)
	}

	mgo := Use(mc.Database("coolcar"))
	mgutil.NewObjectID = primitive.NewObjectID
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
		t.Fatalf("can not create trip: %v", err)
	}

	got, err := mgo.GetTrip(ctx, objid.ToTripId(row.Id), acctId)
	if err != nil {
		t.Errorf("can not get trip: %v", err)
	}

	if diff := cmp.Diff(row, got, protocmp.Transform()); diff != "" {
		t.Errorf("result different; -want +got: %s", diff)
	}
}

func TestMain(m *testing.M) {
	os.Exit(mongotesting.RunWithMongoInDocker(m))
}
