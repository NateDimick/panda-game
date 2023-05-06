package routes

import (
	"encoding/hex"
	"net/http"
	"pandagame/internal/auth"
	"pandagame/internal/mongoconn"
	"time"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

type AuthAPI struct {
	mongo mongoconn.CollectionConn
}

func NewAuthAPI(m mongoconn.CollectionConn) *AuthAPI {
	return &AuthAPI{m}
}

func (a *AuthAPI) LoginUser(w http.ResponseWriter, r *http.Request) {
	uname, pass, ok := r.BasicAuth()
	zap.L().Info("user logging in", zap.String("uname", uname), zap.String("pass", pass))
	if !ok {
		// handle bad basic auth parse
		zap.L().Warn("invalid basic auth provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get record from db by username
	record, err := mongoconn.GetUser(uname, a.mongo)
	if err != nil {
		zap.L().Error("error finding user record to login", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// check password vs db record
	id2, err := hex.DecodeString(record.SecondaryIdentity)
	if err != nil {
		zap.L().Error("error decoding password from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !auth.Verify([]byte(pass), id2, record.PrimaryIdentity) {
		// handle bad login
		zap.L().Warn("password does not match")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c1 := &http.Cookie{
		Name:     "PlayerId",
		Value:    record.TertiaryIdentity,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 8760),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
		MaxAge:   0,
	}
	c2 := &http.Cookie{
		Name:     "UserName",
		Value:    record.Name,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 8760),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
		MaxAge:   0,
	}
	http.SetCookie(w, c1)
	http.SetCookie(w, c2)
	// if checks out, set cookie containing username and playerID
	w.WriteHeader(http.StatusOK)
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
		zap.L().Error("check user exists error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if exists != nil {
		zap.L().Warn("username taken")
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
		zap.L().Error("store new user error", zap.Error(err))
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
