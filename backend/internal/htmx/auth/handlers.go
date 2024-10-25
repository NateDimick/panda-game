package auth

import (
	"net/http"
	"pandagame/internal/config"
	"pandagame/internal/htmx/global"
	"pandagame/internal/pocketbase"
	"pandagame/internal/web"
	"time"

	_ "github.com/a-h/templ"
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
	form := r.PostForm
	username := form.Get("username")
	password := form.Get("password")
	cfg := config.LoadAppConfig()
	resp, err := pocketbase.NewPocketBase(cfg.PB.Address, nil).AsUser().Auth("players").PasswordAuth(pocketbase.AuthPasswordBody{
		Username: username,
		Password: password,
	})
	if err != nil {
		AuthError(err.Error()).Render(r.Context(), w)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:  web.PandaGameCookie,
		Value: resp.Token,
	})
	LoggedInForm(web.IDFromToken(resp.Token)).Render(r.Context(), w)
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
	form := r.PostForm
	username := form.Get("username")
	password := form.Get("password")
	confirmPassword := form.Get("confirmPassword")
	cfg := config.LoadAppConfig()
	_, err := pocketbase.NewPocketBase(cfg.PB.Address, nil).AsUser().Auth("players").Create(pocketbase.NewAuthRecord{
		Credentials: pocketbase.NewAuthCredentials{
			Username:        username,
			Password:        password,
			ConfirmPassword: confirmPassword,
		},
	}, nil)
	if err != nil {
		AuthError(err.Error()).Render(r.Context(), w)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	AfterSignedUp().Render(r.Context(), w)
}
