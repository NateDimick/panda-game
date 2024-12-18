package htmx

import (
	_ "embed"
	"net/http"
	"pandagame/internal/htmx/auth"
	"pandagame/internal/htmx/home"
	"pandagame/internal/htmx/websocket"

	"github.com/go-chi/chi"
)

//go:embed style/output.css
var tailwindcss []byte

func AddHTMXRoutes(r chi.Router) {
	// auth routes
	r.Get("/login", auth.LoginPage)
	r.Get("/signup", auth.SignUpPage)
	r.Get("/logout", auth.LogoutRedirect)
	r.Post("/hmx/signup", auth.ApiSignUp)
	r.Post("/hmx/login", auth.ApiLogin)
	r.Post("/hmx/logout", auth.ApiLogout)
	// main lobby/home page
	r.Get("/", home.ServeHomePage)
	// game routes
	r.Get("/game", websocket.ServeWebsocketUI)
	r.Get("/join", websocket.Join)
	// vanity
	r.Get("/hmx/style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Write(tailwindcss)
	})
}
