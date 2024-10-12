package pocketbase

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecordUnmarshalingMap(t *testing.T) {
	rawJson := "{\"id\": \"hello\", \"@collectionId\": \"five\", \"customAttr\": \"foobar\"}"
	record := new(Record)
	err := json.Unmarshal([]byte(rawJson), record)
	assert.Nil(t, err)
	assert.Equal(t, "hello", record.ID)          // normal record fields are directly unmarshalled
	assert.Equal(t, "five", record.CollectionId) // fields with multiple keys end up in the right place
	cf := record.CustomFields.(map[string]any)
	assert.Equal(t, "foobar", cf["customAttr"]) // custom fields contains additional data
	assert.Equal(t, nil, cf["id"])              // custom fields does not contain known record fields
}

type Example struct {
	CA string `json:"customAttr"`
}

func TestRecordUnmarshalingStruct(t *testing.T) {
	rawJson := "{\"id\": \"hello\", \"@collectionId\": \"five\", \"customAttr\": \"foobar\"}"
	record := new(Record)
	record.CustomFields = new(Example)
	err := json.Unmarshal([]byte(rawJson), record)
	assert.Nil(t, err)
	assert.Equal(t, "hello", record.ID)          // normal record fields are directly unmarshalled
	assert.Equal(t, "five", record.CollectionId) // fields with multiple keys end up in the right place
	e := record.CustomFields.(*Example)
	assert.Equal(t, "foobar", e.CA) // custom fields contains additional data
}

func TestNewRecordMarshal(t *testing.T) {
	r := NewRecord{
		ID: "my-id",
		CustomFields: &Example{
			CA: "my-ca",
		},
	}

	b := bytes.NewBuffer(make([]byte, 0))
	err := json.NewEncoder(b).Encode(r)
	assert.Nil(t, err)
	expectedJSON := "{\"id\":\"my-id\",\"customAttr\":\"my-ca\"}\n"
	assert.Equal(t, expectedJSON, b.String())
}
