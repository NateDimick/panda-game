package pocketbase

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecordUnmarshaling(t *testing.T) {
	rawJson := "{\"id\": \"hello\", \"@collectionId\": \"five\", \"customAttr\": \"foobar\"}"
	record := new(Record)
	err := json.Unmarshal([]byte(rawJson), record)
	assert.Nil(t, err)
	assert.Equal(t, "hello", record.ID)                          // normal record fields are directly unmarshalled
	assert.Equal(t, "five", record.CollectionId)                 // fields with multiple keys end up in the right place
	assert.Equal(t, "foobar", record.CustomFields["customAttr"]) // custom fields contains additional data
	assert.Equal(t, nil, record.CustomFields["id"])              // custom fields does not contain known record fields
}
