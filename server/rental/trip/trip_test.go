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
	"os"
	"testing"
)

func TestCreateTrip(t *testing.T) {
	ctx := auth.ContextWithAccountId(context.Background(), id.AccountId("account1"))
	mc, err := mongotesting.NewClient(ctx)
	if err != nil {
		t.Fatalf("cannot connect to database: %v", err)
	}

	logger, err := server.NewZapLogger()
	if err != nil {
		t.Fatalf("cannot create logger: %v", err)
	}

	pm := &profileManager{}
	cm := &carManager{}

	s := &Service{
		ProfileManager: pm,
		CarManager:     cm,
		PointManager:   &poi.Manager{},
		Mongo:          dao.Use(mc.Database("coolcar")),
		Logger:         logger,
	}

	req := &rentalpb.CreateTripRequest{
		CarId: "car1",
		Start: &rentalpb.Location{
			Latitude:  32.123,
			Longitude: 114.2525,
		},
	}
	pm.identityId = "identity1"
	golden := `{"account_id":"account1","car_id":"car1","start":{"location":{"latitude":32.123,"longitude":114.2525},"point_name":"天安门"},"current":{"location":{"latitude":32.123,"longitude":114.2525},"point_name":"天安门"},"status":1,"identity_id":"identity1"}`

	cases := []struct {
		name         string
		tripId       string
		profileErr   error
		carVerityErr error
		carUnlockErr error
		want         string
		wantErr      bool
	}{
		{
			name:   "normal_create",
			tripId: "61f2b7ef729f5d8b3bd69be4",
			want:   golden,
		},
		{
			name:       "profile_error",
			tripId:     "61f2b7ef729f5d8b3bd69be5",
			profileErr: fmt.Errorf("profile error"),
			wantErr:    true,
		},
		{
			name:         "car_verify_error",
			tripId:       "61f2b7ef729f5d8b3bd69be6",
			carVerityErr: fmt.Errorf("car verify error"),
			wantErr:      true,
		},
		{
			name:         "car_unlock_error",
			tripId:       "61f2b7ef729f5d8b3bd69be7",
			carUnlockErr: fmt.Errorf("car unlock error"),
			wantErr:      false,
			want:         golden,
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			mgutil.NewObjIdWithValue(id.TripId(cc.tripId))
			pm.err = cc.profileErr
			cm.unlockError = cc.carUnlockErr
			cm.verifyError = cc.carVerityErr
			res, err := s.CreateTrip(ctx, req)
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
				t.Errorf("incorrect response: want %q, got %s", cc.want, got)
			}
		})
	}
}

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
}

func (c *carManager) Verify(ctx context.Context, carId id.CarId, location *rentalpb.Location) error {
	return c.verifyError
}
func (c *carManager) Unlock(ctx context.Context, carId id.CarId) error {
	return c.unlockError
}

func TestMain(m *testing.M) {
	os.Exit(mongotesting.RunWithMongoInDocker(m))
}
