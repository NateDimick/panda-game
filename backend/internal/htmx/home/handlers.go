package home

import (
	"net/http"
	"pandagame/internal/htmx/global"
)

func ServeHomePage(w http.ResponseWriter, r *http.Request) {
	// Need to check auth
	token, err := global.IsAuthenticatedRequest(r)
	authenticated := err == nil && token != ""
	global.Page("Panda Game", HomePage(authenticated)).Render(r.Context(), w)
}
