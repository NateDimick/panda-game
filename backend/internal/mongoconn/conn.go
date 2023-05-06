package mongoconn

import (
	"context"
	"fmt"
	"os"
	"pandagame/internal/auth"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func NewMongoConn() CollectionConn {
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		zap.L().Error("")
		panic(err)
	}
	client.Connect(context.TODO())
	client.Ping(context.TODO(), nil)
	return client.Database(os.Getenv("MONGO_DB")).Collection(os.Getenv("MONGO_COLL"))
}

func GetUser(uname string, conn CollectionConn) (*auth.UserRecord, error) {
	result := conn.FindOne(context.TODO(), bson.M{"name": uname})
	if err := result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			zap.L().Info("no user with name " + uname)
			return nil, nil
		} else {
			zap.L().Error("mongo err finding user", zap.Error(err))
			return nil, err
		}
	}
	user := new(auth.UserRecord)
	result.Decode(user)
	zap.L().Info(fmt.Sprintf("found user: %+v", user))
	return user, nil
}

func StoreUser(user *auth.UserRecord, conn CollectionConn) error {
	result, err := conn.InsertOne(context.TODO(), user)
	if err != nil {
		return err
	}
	zap.L().Info("User Stored", zap.String("username", user.Name), zap.String("id", result.InsertedID.(primitive.ObjectID).String()))
	return nil
}
