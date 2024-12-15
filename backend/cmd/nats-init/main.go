package main

import (
	"context"
	"log/slog"
	"os"
	"pandagame/internal/config"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	cfg := config.LoadAppConfig()
	conn, err := nats.Connect(cfg.Nats.Address)
	if err != nil {
		slog.Warn("Failed to connect to nats", slog.String("error", err.Error()))
		os.Exit(1)
	}
	js, err := jetstream.New(conn)
	if err != nil {
		slog.Warn("failed to connect to jetstream", slog.String("error", err.Error()))
		os.Exit(1)
	}
	for status := range js.KeyValueStores(nats.Context(context.Background())).Status() {
		if status.Bucket() == cfg.Nats.GroupBucket {
			slog.Info("jetstream bucket already exists")
			os.Exit(0)
		}
	}
	slog.Info("creating jetstream bucket", slog.String("name", cfg.Nats.GroupBucket))
	_, err = js.CreateKeyValue(nats.Context(context.Background()), jetstream.KeyValueConfig{
		Bucket:      cfg.Nats.GroupBucket,
		TTL:         time.Hour * 240,
		Description: "Panda Game framework group bucket",
		History:     3,
		Storage:     jetstream.MemoryStorage,
	})
	if err != nil {
		slog.Warn("failed to create jetstream bucket", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("created jetstream bucket", slog.String("name", cfg.Nats.GroupBucket))
}
