package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"pandagame/internal/htmx/global"
	"pandagame/internal/sign"
	"pandagame/internal/web"
	"time"

	_ "github.com/a-h/templ"
)

// /login
func LoginPage(w http.ResponseWriter, r *http.Request) {
	token, err := global.IsAuthenticatedRequest(r)
	if err != nil {
		slog.Warn("login: auth check error", slog.String("error", err.Error()))
	}
	authenticated := err == nil && token != ""
	global.Page("Login", Login(authenticated, web.IDFromToken(token))).Render(r.Context(), w)
}

// /signup
func SignUpPage(w http.ResponseWriter, r *http.Request) {
	token, err := global.IsAuthenticatedRequest(r)
	if err != nil {
		slog.Warn("signup: auth check error", slog.String("error", err.Error()))
	}
	authenticated := err == nil && token != ""
	global.Page("Sign Up", SignUp(authenticated, web.IDFromToken(token))).Render(r.Context(), w)
}

// /logout
func LogoutRedirect(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     web.PandaGameCookie,
		Expires:  time.Now().UTC().Add(-time.Hour * 72),
		Value:    "",
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
	})
	http.Redirect(w, r, "/", http.StatusFound)
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
	token, err := sign.In(username, password)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		AuthError(err.Error()).Render(r.Context(), w)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     web.PandaGameCookie,
		Value:    token,
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
	})
	LoggedInForm(web.IDFromToken(token)).Render(r.Context(), w)
}

// /hmx/logout
func ApiLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     web.PandaGameCookie,
		Expires:  time.Now().UTC().Add(-time.Hour * 72),
		Value:    "",
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
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
		w.WriteHeader(http.StatusBadRequest)
		AuthError(err.Error()).Render(r.Context(), w)
		return
	}

	if err := sign.Up(username, password); err != nil {
		slog.Warn("Could not sign up new user", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		AuthError(err.Error()).Render(r.Context(), w)
		return
	}
	AfterSignedUp().Render(r.Context(), w)
}
