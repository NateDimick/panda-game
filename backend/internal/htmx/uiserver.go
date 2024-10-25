package htmx

import (
	"pandagame/internal/htmx/auth"
	"pandagame/internal/htmx/home"
	"pandagame/internal/htmx/websocket"

	"github.com/go-chi/chi"
)

func AddHTMXRoutes(r chi.Router) {
	// auth routes
	r.Get("/login", auth.LoginPage)
	r.Get("/signup", auth.SignUpPage)
	r.Post("/hmx/signup", auth.ApiSignUp)
	r.Post("/hmx/login", auth.ApiLogin)
	r.Post("/hmx/logout", auth.ApiLogout)
	// main lobby/home page
	r.Get("/", home.ServeHomePage)
	// game routes
	r.Get("/game", websocket.ServeWebsocketUI)
	r.Get("/join", websocket.Join)
}
