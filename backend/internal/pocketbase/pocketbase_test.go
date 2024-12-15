package pocketbase

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPrepareRequest(t *testing.T) {
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

	req, err := prepareRequest("POST", "http://localhost:3333/test", r, &tokenHolder{token: "foo", refresher: &tokenRefresher{refreshTime: time.Now().Add(time.Minute)}})
	assert.Nil(t, err)
	body, _ := io.ReadAll(req.Body)
	expectedJSON := "{\"id\":\"some-id\",\"username\":\"test-dude\",\"password\":\"amazing\",\"passwordConfirm\":\"amazing\",\"alpha\":\"beta\",\"gamma\":true}\n"

	assert.Equal(t, expectedJSON, string(body))
}
