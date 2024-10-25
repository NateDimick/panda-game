package main

import (
	"log/slog"
	"net/http"
	"pandagame/internal/config"
	"pandagame/internal/engine"
	"pandagame/internal/framework"
	"pandagame/internal/htmx"
	"pandagame/internal/scaling"
	"pandagame/internal/web"

	"github.com/go-chi/chi"
)

func main() {
	config.SetLogger("panda-server.log")
	appConfig := config.LoadAppConfig()
	fw := framework.NewFramework(&engine.PandaGameEngine{})
	fw.Configure(func(fc *framework.FrameworkConfig) {
		// TODO
		fc.Groups = scaling.Grouper(appConfig)
		fc.Relayer = scaling.Relayer(appConfig)
		fc.IdGenerator = web.IDFromRequest
		fc.Deserializer = engine.MessageDeserializer
		fc.Serializer = engine.MessageSerializer
		// error handler
		fc.ConnectHandler = engine.ConnectionAuthValidator
		// disconnect handler - maybe unneeded?
	})
	mux := chi.NewMux()
	mux.Get("/wss/{type}", fw.ServeHTTP)
	mux.Get("/wss", fw.ServeHTTP)
	htmx.AddHTMXRoutes(mux)
	slog.Info("panda game server is running")
	http.ListenAndServe(":3000", mux)
}
