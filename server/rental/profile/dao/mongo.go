package dao

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	accountIdField      = "accountid"
	profileField        = "profile"
	identityStatusField = profileField + ".identitystatus"
)

type Mongo struct {
	col         *mongo.Collection
	newObjectId func() primitive.ObjectID
}

func Use(db *mongo.Database) *Mongo {
	return &Mongo{
		col:         db.Collection("profile"),
		newObjectId: primitive.NewObjectID,
	}
}

type ProfileRecord struct {
	AccountId string            `bson:"accountid"`
	Profile   *rentalpb.Profile `bson:"profile"`
}

func (m *Mongo) GetProfile(c context.Context, accountId id.AccountId) (*rentalpb.Profile, error) {
	res := m.col.FindOne(c, byAccountId(accountId))

	if err := res.Err(); err != nil {
		return nil, err
	}

	var pr ProfileRecord
	err := res.Decode(&pr)
	if err != nil {
		return nil, fmt.Errorf("cannot decode profile record: %v", err)
	}

	return pr.Profile, nil
}

func (m *Mongo) UpdateProfile(c context.Context, accountId id.AccountId, prevStatus rentalpb.IdentityStatus, profile *rentalpb.Profile) error {
	_, err := m.col.UpdateOne(
		c,
		bson.M{
			accountIdField:      accountId.String(),
			identityStatusField: prevStatus,
		},
		mgutil.Set(bson.M{
			accountIdField: accountId.String(),
			profileField:   profile,
		}),
		options.Update().SetUpsert(true),
	)

	return err
}

func byAccountId(accountId id.AccountId) bson.M {
	return bson.M{
		accountIdField: accountId.String(),
	}
}
