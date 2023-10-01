package redisconn

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

type TestThing struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestGetThing(t *testing.T) {
	testKey := "keyValue"
	testCtx := context.Background()
	mockConn := NewMockRedisConn(t)

	retVal := new(redis.StringCmd)
	retVal.SetVal("{\"name\": \"test\", \"value\": 4}")

	mockConn.EXPECT().Get(testCtx, testKey).Return(retVal)

	thing, err := GetThing[TestThing](testKey, mockConn)
	assert.Nil(t, err)
	assert.Equal(t, "test", thing.Name)
	assert.Equal(t, 4, thing.Value)

}

func TestSetThing(t *testing.T) {
	testKey := "keyValue"
	testCtx := context.Background()
	mockConn := NewMockRedisConn(t)

	thing := &TestThing{
		Name:  "setThing",
		Value: 7,
	}

	stringThing := "{\"name\":\"setThing\",\"value\":7}"

	mockConn.EXPECT().Set(testCtx, testKey, stringThing, time.Hour).Return(new(redis.StatusCmd))

	err := SetThing(testKey, thing, mockConn)

	assert.Nil(t, err)
}
