package config

import (
	"errors"
	"os"

	"github.com/surrealdb/surrealdb.go"
)

type AppConfig struct {
	Surreal       SurrealConfig
	Nats          NatsConfig
	Scale         ScalingLevel
	surrealClient *surrealdb.DB
}

type ScalingLevel int

const (
	Singleton   ScalingLevel = iota // just one instance. framework can operate in memory
	Colocated                       // multiple instances in the same cluster. redis rooms and pocketbase relay
	Distributed                     // unsupported. would require a different db solution than pocketbase
)

var globalConfig *AppConfig

func LoadAppConfig() AppConfig {
	if globalConfig != nil {
		return *globalConfig
	}
	globalConfig = new(AppConfig)
	globalConfig.Surreal.Address = os.Getenv("DB_ADDR")
	globalConfig.Surreal.AdminIdentity = os.Getenv("DB_ADMIN")
	globalConfig.Surreal.AdminPassword = os.Getenv("DB_PASS")
	globalConfig.Surreal.Namespace = os.Getenv("DB_NS")
	globalConfig.Surreal.Database = os.Getenv("DB_DB")
	globalConfig.Nats.Address = os.Getenv("NATS_ADDR")
	globalConfig.Nats.RelaySubject = os.Getenv("NATS_RELAY_SUBJECT")
	globalConfig.Nats.GroupBucket = os.Getenv("NATS_GROUP_BUCKET")
	globalConfig.Scale = loadScaleLevel()

	return *globalConfig
}

func loadScaleLevel() ScalingLevel {
	val := os.Getenv("SCALE")
	switch val {
	case "COLOCATED":
		return Colocated
	case "DISTRIBUTED":
		return Distributed
	case "SINGLETON":
		fallthrough
	default:
		return Singleton
	}
}

func AdminSurreal() (*surrealdb.DB, error) {
	if globalConfig.surrealClient != nil {
		return globalConfig.surrealClient, nil
	}
	db, err := Surreal()
	_, err2 := db.SignIn(&surrealdb.Auth{
		Username: globalConfig.Surreal.AdminIdentity,
		Password: globalConfig.Surreal.AdminPassword,
	})
	globalConfig.surrealClient = db
	return db, errors.Join(err, err2)
}

func Surreal() (*surrealdb.DB, error) {
	db, err := surrealdb.New(globalConfig.Surreal.Address)
	err2 := db.Use(globalConfig.Surreal.Namespace, globalConfig.Surreal.Database)
	return db, errors.Join(err, err2)
}
