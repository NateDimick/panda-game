package routes

import (
	"encoding/hex"
	"log/slog"
	"net/http"
	"pandagame/internal/auth"
	"pandagame/internal/mongoconn"
	"pandagame/internal/redisconn"
	"time"

	"github.com/gofrs/uuid"
)

const sessionPfx string = "s-" // session redis key prefix

type AuthAPI struct {
	mongo mongoconn.CollectionConn
	redis redisconn.RedisConn
}

func NewAuthAPI(m mongoconn.CollectionConn, r redisconn.RedisConn) *AuthAPI {
	return &AuthAPI{m, r}
}

func (a *AuthAPI) LoginUser(w http.ResponseWriter, r *http.Request) {
	uname, pass, ok := r.BasicAuth()
	slog.Info("user logging in", slog.String("uname", uname), slog.String("pass", pass))
	if !ok {
		// handle bad basic auth parse
		slog.Warn("invalid basic auth provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get record from db by username
	record, err := mongoconn.GetUser(uname, a.mongo)
	if err != nil {
		slog.Error("error finding user record to login", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// check password vs db record
	id2, err := hex.DecodeString(record.SecondaryIdentity)
	if err != nil {
		slog.Error("error decoding password from db", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !auth.Verify([]byte(pass), id2, record.PrimaryIdentity) {
		// handle bad login
		slog.Warn("password does not match")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	setSession(w, *record, a.redis)
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
		slog.Error("check user exists error", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if exists != nil {
		slog.Warn("username taken")
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
		Empowered:         true,
	}

	err = mongoconn.StoreUser(record, a.mongo)
	if err != nil {
		slog.Error("store new user error", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// return response
	w.WriteHeader(http.StatusCreated) // actually, should probably redirect
}

func (a *AuthAPI) LoginAsGuest(w http.ResponseWriter, r *http.Request) {
	uname, pass, ok := r.BasicAuth()
	slog.Info("guest user logging in", slog.String("uname", uname), slog.String("pass", pass))
	if !ok {
		// handle bad basic auth parse
		slog.Warn("invalid basic auth provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ensure guest is not taking someone else's name
	record, err := mongoconn.GetUser(uname, a.mongo)
	if err != nil {
		slog.Error("error finding user record to login", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !record.Guest && record.SecondaryIdentity != pass {
		// conflict! username taken by a real user or other guest
		w.WriteHeader(http.StatusConflict)
		return
	} else if record.Guest && record.SecondaryIdentity == pass {
		// guest re-logged in before they expired - re-issue their id
		setSession(w, *record, a.redis)
		w.WriteHeader(http.StatusOK)
		return
	}

	// store record, but with a timeout
	id3 := uuid.Must(uuid.NewV4()).String()
	tempUser := auth.UserRecord{
		Name:              uname,
		PrimaryIdentity:   "???", // should do more here to make guests look like real users
		SecondaryIdentity: pass,
		TertiaryIdentity:  id3,
		Empowered:         false,
		Guest:             true,
		ExpireAt:          time.Now().Add(time.Hour * 24 * 7), // guest expires in one week
	}

	mongoconn.StoreUser(&tempUser, a.mongo)

	// issue cookies
	setSession(w, tempUser, a.redis)
	w.WriteHeader(http.StatusOK)
	return
}

func setSession(w http.ResponseWriter, record auth.UserRecord, redis redisconn.RedisConn) {
	userSession := auth.UserSession{
		SessionID: uuid.Must(uuid.NewV4()).String(),
		Name:      record.Name,
		PlayerID:  record.TertiaryIdentity,
		ExpireAt:  time.Now().Add(time.Hour * 24 * 7), // sessions can last 1 week
	}
	// store session
	redisconn.SetThing(sessionPfx+userSession.SessionID, &userSession, redis)
	// write session id cookie
	sessionCookie := http.Cookie{
		Name:     "pandaGameSession",
		Value:    userSession.SessionID,
		Expires:  userSession.ExpireAt,
		HttpOnly: true,
		Secure:   true,
	}
	http.SetCookie(w, &sessionCookie)
}
