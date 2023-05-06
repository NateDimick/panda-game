package routes

import (
	"net/http"
	"net/http/httptest"
	"pandagame/internal/auth"
	"pandagame/internal/mongoconn"
	"testing"
	_ "unsafe"

	"github.com/stretchr/testify/assert"
)

//go:linkname mockGetUser pandagame/internal/mongoconn.GetUser
func mockGetUser(name string, conn mongoconn.CollectionConn) (*auth.UserRecord, error) {
	return &auth.UserRecord{
		Name:              "nate",
		PrimaryIdentity:   "680065006c006c006f00e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		SecondaryIdentity: "00000000007365742074",
		TertiaryIdentity:  "cb7b4154-1a4d-479b-8930-ff64b838eb5f",
		Empowered:         true,
		Guest:             false,
	}, nil

}

func TestLogin(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/login", nil)
	r.SetBasicAuth("nate", "hello")

	api := new(AuthAPI)

	api.LoginUser(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

}
