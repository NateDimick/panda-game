package mongoconn

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"pandagame/internal/auth"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoConn() CollectionConn {
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		slog.Error("")
		panic(err)
	}
	client.Connect(context.TODO())
	client.Ping(context.TODO(), nil)
	return client.Database(os.Getenv("MONGO_DB")).Collection(os.Getenv("MONGO_COLL"))
}

func GetUser(uname string, conn CollectionConn) (*auth.UserRecord, error) {
	result := conn.FindOne(context.Background(), bson.M{"name": uname})
	if err := result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			slog.Info("no user with name " + uname)
			return nil, nil
		} else {
			slog.Error("mongo err finding user", slog.String("error", err.Error()))
			return nil, err
		}
	}
	user := new(auth.UserRecord)
	result.Decode(user)
	slog.Info(fmt.Sprintf("found user: %+v", user))
	return user, nil
}

func StoreUser(user *auth.UserRecord, conn CollectionConn) error {
	result, err := conn.InsertOne(context.Background(), user)
	if err != nil {
		return err
	}
	slog.Info("User Stored", slog.String("username", user.Name), slog.String("id", result.InsertedID.(primitive.ObjectID).String()))
	return nil
}
