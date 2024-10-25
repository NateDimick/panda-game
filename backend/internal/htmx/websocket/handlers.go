package websocket

import (
	"fmt"
	"net/http"
	"pandagame/internal/htmx/global"
)

func ServeWebsocketUI(w http.ResponseWriter, r *http.Request) {
	// game id will be set as a url fragment in the frontend
	global.Page("Panda Game On!", WSFrame()).Render(r.Context(), w)
}

func Join(w http.ResponseWriter, r *http.Request) {
	gameId := r.URL.Query().Get("gameId")
	w.Header().Set("Location", fmt.Sprintf("/game#%s", gameId))
	w.WriteHeader(http.StatusFound)
}
