package routes

import (
	"encoding/hex"
	"net/http"
	"pandagame/internal/auth"
	"pandagame/internal/mongoconn"
	"strings"

	"github.com/gofrs/uuid"
)

type AuthAPI struct {
	mongo mongoconn.CollectionConn
}

func NewAuthAPI(m mongoconn.CollectionConn) *AuthAPI {
	return &AuthAPI{m}
}

func (a *AuthAPI) LoginUser(w http.ResponseWriter, r *http.Request) {
	uname, pass, ok := r.BasicAuth()
	if !ok {
		// handle bad basic auth parse
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get record from db by username
	record, err := mongoconn.GetUser(uname, a.mongo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// check password vs db record
	id2 := make([]byte, 0)
	_, err = hex.NewDecoder(strings.NewReader(record.SecondaryIdentity)).Read(id2)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !auth.Verify([]byte(pass), id2, record.PrimaryIdentity) {
		// handle bad login
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// if checks out, set cookie containing username and playerID
	w.WriteHeader(http.StatusOK)
	c1 := &http.Cookie{
		Name:  "PlayerId",
		Value: record.TertiaryIdentity,
	}
	c2 := &http.Cookie{
		Name:  "UserName",
		Value: record.Name,
	}
	http.SetCookie(w, c1)
	http.SetCookie(w, c2)
}

func (a *AuthAPI) RegisterUser(w http.ResponseWriter, r *http.Request) {
	// get requested uname and password from request
	uname := r.FormValue("username")
	pass := r.FormValue("password")
	// generate salt, salted password, player id
	id1, id2 := auth.NewSalt([]byte(uname), []byte(pass))
	id3 := uuid.Must(uuid.NewV4()).String()

	// check if username is available
	exists, err := mongoconn.GetUser(uname, a.mongo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if exists != nil {
		w.WriteHeader(http.StatusConflict)
		return
	}

	// write new user if able
	record := &auth.UserRecord{
		Name:              uname,
		PrimaryIdentity:   id1,
		SecondaryIdentity: id2,
		TertiaryIdentity:  id3,
		Guest:             false,
		Empowered:         false,
	}

	err = mongoconn.StoreUser(record, a.mongo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// return response
	w.WriteHeader(http.StatusCreated) // actually, should probably redirect
}

// allows an empowered user to empower other users
// specify username to empower by query param
func (a *AuthAPI) EmpowerUser(w http.ResponseWriter, r *http.Request) {
	// get requester auth

	// check if user is empowered

	// get specified user

	// empower them and updated
}

func (a *AuthAPI) LoginAsGuest(w http.ResponseWriter, r *http.Request) {

}
