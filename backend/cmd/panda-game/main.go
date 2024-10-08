package main

import (
	"log/slog"
	"net/http"
	"pandagame/internal/config"
	"pandagame/internal/engine"
	"pandagame/internal/framework"
	"pandagame/internal/scaling"

	"github.com/go-chi/chi"
)

func main() {
	config.SetLogger()
	appConfig := config.LoadAppConfig()
	fw := framework.NewFramework(&engine.PandaGameEngine{})
	fw.Configure(func(fc *framework.FrameworkConfig) {
		// TODO
		fc.Groups = scaling.Grouper(appConfig)
		fc.Relayer = scaling.Relayer(appConfig)
		fc.IdGenerator = engine.IDFromToken
		fc.Deserializer = engine.MessageDeserializer
		fc.Serializer = engine.MessageSerializer
		// error handler
		fc.ConnectHandler = engine.ConnectionAuthValidator
		// disconnect handler - maybe unneeded?
	})
	mux := chi.NewMux()
	mux.Get("/wss", fw.ServeHTTP)
	slog.Info("panda game server is running")
	http.ListenAndServe(":3000", mux)
}
