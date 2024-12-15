package home

import (
	"log/slog"
	"net/http"
	"pandagame/internal/htmx/global"
	"pandagame/internal/web"
)

func ServeHomePage(w http.ResponseWriter, r *http.Request) {
	// Need to check auth
	token, err := global.IsAuthenticatedRequest(r)
	if err != nil {
		slog.Warn("home: auth check error", slog.String("error", err.Error()))
	}
	authenticated := err == nil && token != ""
	username := web.IDFromToken(token)
	global.Page("Panda Game", HomePage(authenticated, username)).Render(r.Context(), w)
}
