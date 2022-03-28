package mgutil

import (
	"coolcar/shared/mongo/objid"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	IdFieldName        = "_id"
	UpdatedAtFieldName = "updatedat"
)

type IdField struct {
	Id primitive.ObjectID `bson:"_id"`
}

type UpdatedAtField struct {
	UpdatedAt int64 `bson:"updatedat"`
}

var NewObjectId = primitive.NewObjectID

func NewObjIdWithValue(id fmt.Stringer) {
	NewObjectId = func () primitive.ObjectID {
		return objid.EnsureObjId(id)
	}
}
var UpdatedAt = func() int64 {
	return time.Now().UnixNano()
}
