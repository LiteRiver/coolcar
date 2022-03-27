package mgutil

import (
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

var NewObjectID = primitive.NewObjectID
var UpdatedAt = func() int64 {
	return time.Now().UnixNano()
}
