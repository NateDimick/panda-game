package auth

import (
	"net/http"
	"pandagame/internal/htmx/global"

	_ "github.com/a-h/templ"
)

// /login
func LoginPage(w http.ResponseWriter, r *http.Request) {
	// TODO: get auth cookie, check if user is logged in
	global.Page("Login", Login(false, "")).Render(r.Context(), w)
}

// /signup
func SignUpPage(w http.ResponseWriter, r *http.Request) {
	// TODO get auth cookie, yada yada
	global.Page("Sign Up", SignUp(false, "")).Render(r.Context(), w)
}

// /hmx/login
func ApiLogin(w http.ResponseWriter, r *http.Request) {
	// TODO - open form, login with credentials, blah blah
	// TODO - if error, return code 400 with AuthError
	LoggedInForm("todo").Render(r.Context(), w)
}

// /hmx/logout
func ApiLogout(w http.ResponseWriter, r *http.Request) {
	// TODO: strip cookies, blah blah blah
	LoginForm().Render(r.Context(), w)
}

// /hmx/signup
func ApiSignUp(w http.ResponseWriter, r *http.Request) {
	// TODO - do pocketbase user creation
	// TODO - if error, return code 400 with AuthError
	AfterSignedUp().Render(r.Context(), w)
}
