package main

import (
	"log/slog"
	"net/http"
	"pandagame/internal/config"
	"pandagame/internal/engine"
	"pandagame/internal/framework"

	"github.com/go-chi/chi"
)

func main() {
	config.SetLogger()
	fw := framework.NewFramework(&engine.PandaGameEngine{})
	fw.Configure(func(fc *framework.FrameworkConfig) {
		// TODO
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
