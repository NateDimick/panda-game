package pocketbase

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuthRecordMarshal(t *testing.T) {
	ar := &NewAuthRecord{
		Credentials: NewAuthCredentials{
			Username:        "test-user",
			Password:        "pw",
			ConfirmPassword: "pw",
		},
		CustomFields: map[string]any{
			"one": "value",
			"two": "more value",
		},
	}

	b, err := json.Marshal(ar)
	assert.Nil(t, err)
	out := make(map[string]string)
	err = json.Unmarshal(b, &out)
	assert.Nil(t, err)
	assert.Equal(t, 5, len(out))
	assert.Equal(t, "value", out["one"])
	assert.Equal(t, "pw", out["password"])
}

func TestGetClaims(t *testing.T) {
	adminToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mjk3MzY0NDMsImlkIjoibzhsdHFlM2tjd25lOTd6IiwidHlwZSI6ImFkbWluIn0.vyRB-bz7cb2l1ha-1A34p_gOX_jIVOVqjAGxtExODm8"
	userToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2xsZWN0aW9uSWQiOiJfcGJfdXNlcnNfYXV0aF8iLCJleHAiOjE3Mjk5MTExNDUsImlkIjoidWR4ODJyNjV6Zm12MjJwIiwidHlwZSI6ImF1dGhSZWNvcmQifQ.PM4Z5Ai92dr_NGsNqbiQMMrmeAolx9O1y5B-LdkzxsM"

	assert.Equal(t, "", getCollectionID(adminToken))
	assert.Equal(t, "_pb_users_auth_", getCollectionID(userToken))
	assert.True(t, time.Date(2025, time.August, 27, 5, 0, 0, 0, time.UTC).After(getExpiryTime(adminToken)))
	assert.True(t, time.Date(2024, time.October, 11, 5, 0, 0, 0, time.UTC).Before(getExpiryTime(adminToken)))
}
