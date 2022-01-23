package dao

import (
	"context"
	mongotesting "coolcar/shared/mongo/testing"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoURI string

func TestGetAccountId(t *testing.T) {
	ctx := context.Background()
	mc, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Fatalf("cannot connect to database: %v", err)
	}

	mgo := Use(mc.Database("coolcar"))

	_, err = mgo.col.InsertMany(ctx, []interface{}{
		bson.M{
			"_id":     mustObjId("61e6f5f063f1d007f671b034"),
			"open_id": "open_id_1",
		},
		bson.M{
			"_id":     mustObjId("61e6f5f063f1d007f671b027"),
			"open_id": "open_id_2",
		},
	})

	if err != nil {
		t.Fatalf("cannot insert initial values: %v", err)
	}

	mgo.newObjectId = func() primitive.ObjectID {
		objId, _ := primitive.ObjectIDFromHex("61e6f5f063f1d007f671b022")
		return objId
	}

	cases := []struct {
		name   string
		openId string
		want   string
	}{
		{
			name:   "existing_user",
			openId: "open_id_1",
			want:   "61e6f5f063f1d007f671b034",
		},
		{
			name:   "another_existing_user",
			openId: "open_id_2",
			want:   "61e6f5f063f1d007f671b027",
		},
		{
			name:   "new_user",
			openId: "open_id_3",
			want:   "61e6f5f063f1d007f671b022",
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			id, err := mgo.GetAccountId(context.Background(), cs.openId)
			if err != nil {
				t.Errorf("failed to getAccountId: %v", err)
			}

			if id != cs.want {
				t.Errorf("getAccountId, want: %q, got: %q", cs.want, id)
			}
		})
	}
}

func mustObjId(hex string) primitive.ObjectID {
	objId, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		panic(err)
	}

	return objId
}

func TestMain(m *testing.M) {
	os.Exit(mongotesting.RunWithMongoInDocker(m, &mongoURI))
}
