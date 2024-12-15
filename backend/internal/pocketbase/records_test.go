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

func jsonMarsh(j any) string {
	if j != nil {
		bb := bytes.NewBuffer(make([]byte, 0))
		json.NewEncoder(bb).Encode(j)
		return bb.String()
	}
	return ""
}

func TestNewAuthRecordMarshal(t *testing.T) {
	r := NewRecord{
		ID: "some-id",
		CustomFields: &NewAuthRecord{
			Credentials: NewAuthCredentials{
				Username:        "test-dude",
				Password:        "amazing",
				ConfirmPassword: "amazing",
			},
			CustomFields: map[string]any{
				"alpha": "beta",
				"gamma": true,
			},
		},
	}
	b := bytes.NewBuffer(make([]byte, 0))
	err := json.NewEncoder(b).Encode(r)
	assert.Nil(t, err)
	expectedJSON := "{\"id\":\"some-id\",\"username\":\"test-dude\",\"password\":\"amazing\",\"passwordConfirm\":\"amazing\",\"alpha\":\"beta\",\"gamma\":true}\n"
	assert.Equal(t, expectedJSON, b.String())

	json2 := jsonMarsh(r)
	assert.Equal(t, expectedJSON, json2)
}
