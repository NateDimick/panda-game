package routes

import (
	"encoding/hex"
	"net/http"
	"pandagame/internal/auth"
	"pandagame/internal/mongoconn"
	"pandagame/internal/redisconn"
	"time"

	"github.com/gofrs/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const sessionSfx string = "-session"

type AuthAPI struct {
	mongo mongoconn.CollectionConn
	redis redisconn.RedisConn
}

func NewAuthAPI(m mongoconn.CollectionConn, r redisconn.RedisConn) *AuthAPI {
	return &AuthAPI{m, r}
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
	// get session cookie
	sessionCookie, err := r.Cookie("pandaGameSession")
	if err == http.ErrNoCookie {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Absent session, please log in"))
		return
	}
	sessionID := sessionCookie.Value

	// get session info
	session, err := redisconn.GetThing[auth.UserSession](sessionID+sessionSfx, a.redis)
	if err == redis.Nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Session expired, please log back in"))
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !session.Empowered {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// get specified user
	userToEmpower := r.URL.Query().Get("username")
	user, err := mongoconn.GetUser(userToEmpower, a.mongo)
	if err != nil {
		zap.L().Error("error finding user record to empower", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// empower them and update
	user.Empowered = true
	if err := mongoconn.StoreUser(user, a.mongo); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (a *AuthAPI) LoginAsGuest(w http.ResponseWriter, r *http.Request) {
	uname, pass, ok := r.BasicAuth()
	zap.L().Info("guest user logging in", zap.String("uname", uname), zap.String("pass", pass))
	if !ok {
		// handle bad basic auth parse
		zap.L().Warn("invalid basic auth provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ensure guest is not taking someone else's name
	record, err := mongoconn.GetUser(uname, a.mongo)
	if err != nil {
		zap.L().Error("error finding user record to login", zap.Error(err))
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
	redisconn.SetThing(userSession.SessionID+sessionSfx, &userSession, redis)
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
