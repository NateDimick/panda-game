package mongoconn

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CollectionConn interface {
	FindOne(context.Context, interface{}, ...*options.FindOneOptions) *mongo.SingleResult
	InsertOne(context.Context, interface{}, ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
}
