package mongoconn

import (
	"context"
	"pandagame/internal/auth"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// func TestGetUser(t *testing.T) {
// 	mockCollection := NewMockCollectionConn(t)
// 	mockOptions := bson.M{"name": "test-user"}
// 	mockResult := new(mongo.SingleResult)
// 	mockResult. // no way to set the raw value of the result, no way to test
// 	mockCollection.EXPECT().FindOne(context.Background(), mockOptions).Return(mockResult)

// 	record, err := GetUser("test-user", mockCollection)
// 	assert.Nil(t, err)
// }

func TestStoreUser(t *testing.T) {
	mockCollection := NewMockCollectionConn(t)
	testRecord := &auth.UserRecord{
		Name: "test-user",
	}
	mockResult := new(mongo.InsertOneResult)
	mockResult.InsertedID = primitive.ObjectID([12]byte([]byte("abcdabcdabcd")))
	mockCollection.EXPECT().InsertOne(context.Background(), testRecord).Return(mockResult, nil)

	err := StoreUser(testRecord, mockCollection)
	assert.Nil(t, err)
}
