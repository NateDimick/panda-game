package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"pandagame/internal/config"
	"pandagame/internal/htmx/global"
	"pandagame/internal/web"
	"time"

	_ "github.com/a-h/templ"
	"github.com/surrealdb/surrealdb.go"
)

// /login
func LoginPage(w http.ResponseWriter, r *http.Request) {
	token, err := global.IsAuthenticatedRequest(r)
	authenticated := err == nil && token != ""
	global.Page("Login", Login(authenticated, web.IDFromToken(token))).Render(r.Context(), w)
}

// /signup
func SignUpPage(w http.ResponseWriter, r *http.Request) {
	token, err := global.IsAuthenticatedRequest(r)
	authenticated := err == nil && token != ""
	global.Page("Sign Up", SignUp(authenticated, web.IDFromToken(token))).Render(r.Context(), w)
}

// /hmx/login
func ApiLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		AuthError(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	form := r.PostForm
	username := form.Get("username")
	password := form.Get("password")
	db, _ := config.Surreal()
	token, err := db.SignIn(&surrealdb.Auth{
		Username: username,
		Password: password,
		Access:   "user",
	})

	if err != nil {
		AuthError(err.Error()).Render(r.Context(), w)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:  web.PandaGameCookie,
		Value: token,
	})
	LoggedInForm(web.IDFromToken(token)).Render(r.Context(), w)
}

// /hmx/logout
func ApiLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    web.PandaGameCookie,
		Expires: time.Now().UTC().Add(-time.Hour * 72),
	})
	LoginForm().Render(r.Context(), w)
}

// /hmx/signup
func ApiSignUp(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		AuthError(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	form := r.PostForm
	username := form.Get("username")
	password := form.Get("password")
	confirmPassword := form.Get("confirmPassword")
	if password != confirmPassword {
		err := errors.New("passwords do not match")
		AuthError(err.Error()).Render(r.Context(), w)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	db, _ := config.Surreal()
	_, err := db.SignUp(&surrealdb.Auth{
		Username: username,
		Password: password,
		Access:   "user",
	})
	if err != nil {
		slog.Warn("Could not sign up new user", slog.String("error", err.Error()))
		AuthError(err.Error()).Render(r.Context(), w)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	AfterSignedUp().Render(r.Context(), w)
}
