package config

import (
	"os"
	"pandagame/internal/pocketbase"
)

type AppConfig struct {
	PB            PocketBaseConfig
	Nats          NatsConfig
	Scale         ScalingLevel
	pbAdminClient *pocketbase.PBClient
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
	globalConfig.PB.Address = os.Getenv("PB_ADDR")
	globalConfig.PB.AdminIdentity = os.Getenv("PB_ADMIN")
	globalConfig.PB.AdminPassword = os.Getenv("PB_ADMIN_PASS")
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

func PBAdmin() pocketbase.AdminAPI {
	if globalConfig.pbAdminClient == nil {
		globalConfig.pbAdminClient = pocketbase.NewPocketBase(globalConfig.PB.Address, nil)
		globalConfig.pbAdminClient.AsAdmin().Admins().PasswordAuth(pocketbase.AdminPasswordBody{
			Identity: globalConfig.PB.AdminIdentity,
			Password: globalConfig.PB.AdminPassword,
		})
	}
	return globalConfig.pbAdminClient.AsAdmin()
}
