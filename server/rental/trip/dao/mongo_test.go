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
		mgutil.NewObjIdWithValue(id.TripId(cc.tripId))

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
	mgutil.NewObjectId = primitive.NewObjectID
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

func TestGetTrips(t *testing.T) {
	ctx := context.Background()
	mc, err := mongotesting.NewClient(ctx)
	if err != nil {
		t.Fatalf("cannot connect to database: %v", err)
	}

	mgo := Use(mc.Database("coolcar"))

	rows := []struct {
		id        string
		accountId id.AccountId
		status    rentalpb.TripStatus
	}{
		{
			id:        "61f2b7ef729f5d8b3bc69be4",
			accountId: "account_id_for_get_trips",
			status:    rentalpb.TripStatus_FINISHED,
		},
		{
			id:        "61f2b7ef729f5d8b3bc69be5",
			accountId: "account_id_for_get_trips",
			status:    rentalpb.TripStatus_FINISHED,
		},
		{
			id:        "61f2b7ef729f5d8b3bc69be6",
			accountId: "account_id_for_get_trips",
			status:    rentalpb.TripStatus_FINISHED,
		},
		{
			id:        "61f2b7ef729f5d8b3bc69be7",
			accountId: "account_id_for_get_trips",
			status:    rentalpb.TripStatus_IN_PROGRESS,
		},
		{
			id:        "61f2b7ef729f5d8b3bc69be8",
			accountId: "account_id_for_get_trips_1",
			status:    rentalpb.TripStatus_IN_PROGRESS,
		},
	}

	for _, r := range rows {
		mgutil.NewObjIdWithValue(id.TripId(r.id))
		_, err := mgo.CreateTrip(ctx, &rentalpb.Trip{
			AccountId: r.accountId.String(),
			Status:    r.status,
		})

		if err != nil {
			t.Fatalf("cannot create rows: %v", err)
		}
	}

	cases := []struct {
		name       string
		accountId  string
		status     rentalpb.TripStatus
		wantCount  int
		wantOnlyId string
	}{
		{
			name:      "get_all",
			accountId: "account_id_for_get_trips",
			status:    rentalpb.TripStatus_TS_NOT_SPECIFIED,
			wantCount: 4,
		},
		{
			name:       "get_in_progress",
			accountId:  "account_id_for_get_trips",
			status:     rentalpb.TripStatus_IN_PROGRESS,
			wantCount:  1,
			wantOnlyId: "61f2b7ef729f5d8b3bc69be7",
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			res, err := mgo.GetTrips(context.Background(), id.AccountId(cc.accountId), cc.status)
			if err != nil {
				t.Errorf("cannot get trips: %v", err)
			}

			if cc.wantCount != len(res) {
				t.Errorf("incorrect result count; want: %d, got: %d", cc.wantCount, len(res))
			}

			if cc.wantOnlyId != "" && len(res) > 0 {
				if cc.wantOnlyId != res[0].Id.Hex() {
					t.Errorf("only_id incorrect; want: %q, got: %q", cc.accountId, res[0].Id.Hex())
				}
			}
		})
	}
}

func TestUpdateTrip(t *testing.T) {
	ctx := context.Background()
	mc, err := mongotesting.NewClient(ctx)
	if err != nil {
		t.Fatalf("cannot connect to database: %v", err)
	}

	mgo := Use(mc.Database("coolcar"))
	tripId := id.TripId("61f2b7ef729f5d9b3bc69be7")
	accountId := id.AccountId("account_id_for_update")

	var now int64 = 10000
	mgutil.NewObjIdWithValue(tripId)
	mgutil.UpdatedAt = func() int64 {
		return now
	}

	row, err := mgo.CreateTrip(ctx, &rentalpb.Trip{
		AccountId: accountId.String(),
		Status:    rentalpb.TripStatus_IN_PROGRESS,
		Start: &rentalpb.LocationStatus{
			PointName: "start point",
		},
	})

	if err != nil {
		t.Fatalf("cannot create trip: %v", err)
	}

	if row.UpdatedAt != now {
		t.Fatalf("wrong updatedat; want: 10000, got: %d", row.UpdatedAt)
	}

	update := &rentalpb.Trip{
		AccountId: accountId.String(),
		Status:    rentalpb.TripStatus_IN_PROGRESS,
		Start: &rentalpb.LocationStatus{
			PointName: "start point updated",
		},
	}

	cases := []struct {
		name          string
		now           int64
		withUdpatedAt int64
		wantErr       bool
	}{
		{
			name:          "normal_update",
			now:           20000,
			withUdpatedAt: 10000,
			wantErr:       false,
		},
		{
			name:          "update_with_stale_timestamps",
			now:           30000,
			withUdpatedAt: 10000,
			wantErr:       true,
		},
		{
			name:          "update_with_refetch",
			now:           40000,
			withUdpatedAt: 20000,
			wantErr:       false,
		},
	}

	for _, cc := range cases {
		now = cc.now
		err := mgo.UpdateTrip(ctx, tripId, accountId, cc.withUdpatedAt, update)
		if cc.wantErr {
			if err == nil {
				t.Errorf("%s: want error, got none", cc.name)
			} else {
				continue
			}
		} else {
			if err != nil {
				t.Errorf("%s: cannot update: %v", cc.name, err)
			}
		}

		updatedTrip, err := mgo.GetTrip(ctx, tripId, accountId)
		if err != nil {
			t.Errorf("%s: cannot get updated trip: %v", cc.name, err)
		}

		if cc.now != updatedTrip.UpdatedAt {
			t.Errorf("%s: incorrect updatedat: want: %d, got: %d", cc.name, cc.now, updatedTrip.UpdatedAt)
		}
	}
}

func TestMain(m *testing.M) {
	os.Exit(mongotesting.RunWithMongoInDocker(m))
}
