package profile

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/rental/profile/dao"
	"coolcar/shared/auth"
	"coolcar/shared/id"
	mongotesting "coolcar/shared/mongo/testing"
	"coolcar/shared/server"
	"fmt"
	"os"
	"testing"
)

func TestProfileLifecycle(t *testing.T) {
	ctx := context.Background()
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

	s := &Service{
		Logger: logger,
		Mongo:  dao.Use(db),
	}

	accountId := id.AccountId("account1")
	c := auth.ContextWithAccountId(ctx, accountId)
	cases := []struct {
		name       string
		op         func() (*rentalpb.Profile, error)
		wantName   string
		wantStatus rentalpb.IdentityStatus
		wantErr    bool
	}{
		{
			name: "get_empty",
			op: func() (*rentalpb.Profile, error) {
				return s.GetProfile(c, &rentalpb.GetProfileRequest{})
			},
			wantName:   "",
			wantStatus: rentalpb.IdentityStatus_UNSUBMITTED,
		},
		{
			name: "submit",
			op: func() (*rentalpb.Profile, error) {
				return s.SubmitProfile(c, &rentalpb.Identity{Name: "abc"})
			},
			wantName:   "abc",
			wantStatus: rentalpb.IdentityStatus_PENDING,
		},
		{
			name: "submit_again",
			op: func() (*rentalpb.Profile, error) {
				return s.SubmitProfile(c, &rentalpb.Identity{Name: "abc"})
			},
			wantName: "abc",
			wantErr:  true,
		},
		{
			name: "todo_force_verify",
			op: func() (*rentalpb.Profile, error) {
				p := &rentalpb.Profile{
					Identity: &rentalpb.Identity{
						Name: "abc",
					},
					IdentityStatus: rentalpb.IdentityStatus_VERIFIED,
				}

				err := s.Mongo.UpdateProfile(c, accountId, rentalpb.IdentityStatus_PENDING, p)
				if err != nil {
					return nil, err
				}
				return p, nil
			},
			wantName:   "abc",
			wantStatus: rentalpb.IdentityStatus_VERIFIED,
		},
		{
			name: "clear",
			op: func() (*rentalpb.Profile, error) {
				return s.ClearProfile(c, &rentalpb.ClearProfileRequest{})
			},
			wantName:   "",
			wantStatus: rentalpb.IdentityStatus_UNSUBMITTED,
		},
	}

	for _, cc := range cases {
		fmt.Printf("%s: begin testing\n", cc.name)
		p, err := cc.op()
		if cc.wantErr {
			if err == nil {
				t.Errorf("%s: wnat error; got none", cc.name)
			} else {
				continue
			}
		}

		if err != nil {
			t.Errorf("%s: operation failed: %v", cc.name, err)
			continue
		}

		if cc.wantName != "" && p.Identity.Name != cc.wantName {
			t.Errorf("%s: want name: %s, got name: %s", cc.name, cc.wantName, p.Identity.Name)
			continue
		}

		if p.IdentityStatus != cc.wantStatus {
			t.Errorf("%s: want status: %d, got status: %d", cc.name, cc.wantStatus, p.IdentityStatus)
			continue
		}
	}
}

func TestMain(m *testing.M) {
	os.Exit(mongotesting.RunWithMongoInDocker(m))
}
