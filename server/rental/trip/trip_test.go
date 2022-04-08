package trip

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/rental/trip/client/poi"
	"coolcar/rental/trip/dao"
	"coolcar/shared/auth"
	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"coolcar/shared/mongo/testing"
	"coolcar/shared/server"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"
)

type profileManager struct {
	identityId id.IdentityId
	err        error
}

func (p *profileManager) Verify(ctx context.Context, accountId id.AccountId) (id.IdentityId, error) {
	return p.identityId, p.err
}

type carManager struct {
	verifyError error
	unlockError error
	// lockError   error
}

func (c *carManager) Verify(ctx context.Context, carId id.CarId, location *rentalpb.Location) error {
	return c.verifyError
}

func (c *carManager) Unlock(ctx context.Context, carId id.CarId, accountId id.AccountId, avatarUrl string, tripId id.TripId) error {
	return c.unlockError
}

func (c *carManager) Lock(ctx context.Context, carId id.CarId) error {
	// return c.lockError
	return nil
}

type distanceCalc struct{}

func (d *distanceCalc) DistanceKm(ctx context.Context, from *rentalpb.Location, to *rentalpb.Location) (float64, error) {
	if from.Latitude == to.Latitude && from.Longitude == to.Longitude {
		return 0, nil
	}

	return 100, nil
}

func TestCreateTrip(t *testing.T) {
	ctx := context.Background()
	pm := &profileManager{}
	cm := &carManager{}
	s := newService(ctx, t, pm, cm)
	nowFunc = func() int64 {
		return 1648551144
	}
	req := &rentalpb.CreateTripRequest{
		CarId: "car1",
		Start: &rentalpb.Location{
			Latitude:  32.123,
			Longitude: 114.2525,
		},
	}
	pm.identityId = "identity1"
	golden := `{"account_id":%q,"car_id":"car1","start":{"location":{"latitude":32.123,"longitude":114.2525},"point_name":"天安门","timestamp_sec":1648551144},"current":{"location":{"latitude":32.123,"longitude":114.2525},"point_name":"天安门","timestamp_sec":1648551144},"status":1,"identity_id":"identity1"}`

	cases := []struct {
		name         string
		accountId    string
		tripId       string
		profileErr   error
		carVerityErr error
		carUnlockErr error
		want         string
		wantErr      bool
	}{
		{
			name:      "normal_create",
			accountId: "account1",
			tripId:    "61f2b7ef729f5d8b3bd69be4",
			want:      fmt.Sprintf(golden, "account1"),
		},
		{
			name:       "profile_error",
			accountId:  "account2",
			tripId:     "61f2b7ef729f5d8b3bd69be5",
			profileErr: fmt.Errorf("profile error"),
			wantErr:    true,
		},
		{
			name:         "car_verify_error",
			accountId:    "account3",
			tripId:       "61f2b7ef729f5d8b3bd69be6",
			carVerityErr: fmt.Errorf("car verify error"),
			wantErr:      true,
		},
		{
			name:         "car_unlock_error",
			accountId:    "account4",
			tripId:       "61f2b7ef729f5d8b3bd69be7",
			carUnlockErr: fmt.Errorf("car unlock error"),
			wantErr:      false,
			want:         fmt.Sprintf(golden, "account4"),
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			mgutil.NewObjIdWithValue(id.TripId(cc.tripId))
			pm.err = cc.profileErr
			cm.unlockError = cc.carUnlockErr
			cm.verifyError = cc.carVerityErr
			c := auth.ContextWithAccountId(context.Background(), id.AccountId(cc.accountId))
			res, err := s.CreateTrip(c, req)
			if cc.wantErr {
				if err == nil {
					t.Errorf("want error; got none")
				} else {
					return
				}
			}

			if err != nil {
				t.Errorf("error creating trip: %v", err)
				return
			}
			if res.Id != cc.tripId {
				t.Errorf("incorrect id; want %q, got %q", cc.tripId, res.Id)
			}
			b, err := json.Marshal(res.Trip)
			if err != nil {
				t.Errorf("cannot marshal response: %v", err)
			}

			got := string(b)
			if cc.want != got {
				t.Errorf("incorrect response: want %s, got %s", cc.want, got)
			}
		})
	}
}

func TestTripLifecycle(t *testing.T) {
	ctx := auth.ContextWithAccountId(context.Background(), id.AccountId("account_for_lifecycle"))
	s := newService(ctx, t, &profileManager{}, &carManager{})

	tripId := id.TripId("61f2b7ef729f5d8d3bd69be7")
	mgutil.NewObjIdWithValue(tripId)
	cases := []struct {
		name    string
		now     int64
		op      func() (*rentalpb.Trip, error)
		want    string
		wantErr bool
	}{
		{
			name: "create_trip",
			now:  10000,
			op: func() (*rentalpb.Trip, error) {
				e, err := s.CreateTrip(ctx, &rentalpb.CreateTripRequest{
					CarId: "car1",
					Start: &rentalpb.Location{
						Latitude:  32.123,
						Longitude: 114.2525,
					},
				})
				if err != nil {
					return nil, err
				}
				return e.Trip, nil
			},
			want: `{"account_id":"account_for_lifecycle","car_id":"car1","start":{"location":{"latitude":32.123,"longitude":114.2525},"point_name":"天安门","timestamp_sec":10000},"current":{"location":{"latitude":32.123,"longitude":114.2525},"point_name":"天安门","timestamp_sec":10000},"status":1}`,
		},
		{
			name: "update_trip",
			now:  20000,
			op: func() (*rentalpb.Trip, error) {
				return s.UpdateTrip(ctx, &rentalpb.UpdateTripRequest{
					Id: tripId.String(),
					Current: &rentalpb.Location{
						Latitude:  28.232325,
						Longitude: 123.2343221,
					},
				})
			},
			want: `{"account_id":"account_for_lifecycle","car_id":"car1","start":{"location":{"latitude":32.123,"longitude":114.2525},"point_name":"天安门","timestamp_sec":10000},"current":{"location":{"latitude":28.232325,"longitude":123.2343221},"fee_cent":7968,"km_driven":100,"point_name":"天安门","timestamp_sec":20000},"status":1}`,
		},
		{
			name: "finish_trip",
			now:  30000,
			op: func() (*rentalpb.Trip, error) {
				return s.UpdateTrip(ctx, &rentalpb.UpdateTripRequest{
					Id:      tripId.String(),
					EndTrip: true,
				})
			},
			want: `{"account_id":"account_for_lifecycle","car_id":"car1","start":{"location":{"latitude":32.123,"longitude":114.2525},"point_name":"天安门","timestamp_sec":10000},"current":{"location":{"latitude":28.232325,"longitude":123.2343221},"fee_cent":11825,"km_driven":100,"point_name":"天安门","timestamp_sec":30000},"end":{"location":{"latitude":28.232325,"longitude":123.2343221},"fee_cent":11825,"km_driven":100,"point_name":"天安门","timestamp_sec":30000},"status":2}`,
		},
		{
			name: "query_trip",
			now:  40000,
			op: func() (*rentalpb.Trip, error) {
				return s.GetTrip(ctx, &rentalpb.GetTripRequest{
					Id: tripId.String(),
				})
			},
			want: `{"account_id":"account_for_lifecycle","car_id":"car1","start":{"location":{"latitude":32.123,"longitude":114.2525},"point_name":"天安门","timestamp_sec":10000},"current":{"location":{"latitude":28.232325,"longitude":123.2343221},"fee_cent":11825,"km_driven":100,"point_name":"天安门","timestamp_sec":30000},"end":{"location":{"latitude":28.232325,"longitude":123.2343221},"fee_cent":11825,"km_driven":100,"point_name":"天安门","timestamp_sec":30000},"status":2}`,
		},
		{
			name: "udpate_after_finished",
			now:  50000,
			op: func() (*rentalpb.Trip, error) {
				return s.UpdateTrip(ctx, &rentalpb.UpdateTripRequest{
					Id: tripId.String(),
				})
			},
			wantErr: true,
		},
	}

	rand.Seed(1345)
	for _, cc := range cases {
		nowFunc = func() int64 {
			return cc.now
		}

		trip, err := cc.op()

		if cc.wantErr {
			if err == nil {
				t.Errorf("%s: want error; got none", cc.name)
			} else {
				continue
			}
		}

		if err != nil {
			t.Errorf("%s: operation failed: %v", cc.name, err)
			continue
		}

		b, err := json.Marshal(trip)
		if err != nil {
			t.Errorf("%s: failed to marshal response: %v", cc.name, err)
		}

		got := string(b)
		if cc.want != got {
			t.Errorf("%s: incorrect response; want: %s, got: %s", cc.name, cc.want, got)
		}
	}
}

func newService(ctx context.Context, t *testing.T, pm ProfileManager, cm CarManager) *Service {
	mc, err := mongotesting.NewClient(ctx)
	if err != nil {
		t.Fatalf("cannot connect to database: %v", err)
	}

	logger, err := server.NewZapLogger()
	if err != nil {
		t.Fatalf("cannot create logger: %v", err)
	}
	db := mc.Database("coolcar")
	mongotesting.SetupIndices(ctx, db)

	return &Service{
		ProfileManager: pm,
		CarManager:     cm,
		PointManager:   &poi.Manager{},
		DistanceCalc:   &distanceCalc{},
		Mongo:          dao.Use(db),
		Logger:         logger,
	}
}

func TestMain(m *testing.M) {
	os.Exit(mongotesting.RunWithMongoInDocker(m))
}
