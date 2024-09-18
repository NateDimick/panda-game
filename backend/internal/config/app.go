package config

import "os"

type AppConfig struct {
	PB    PocketBaseConfig
	Redis RedisConfig
	Scale ScalingLevel
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
	globalConfig.Redis.Address = os.Getenv("REDIS_ADDR")
	globalConfig.PB.Address = os.Getenv("PB_ADDR")
	globalConfig.PB.AdminIdentity = os.Getenv("PB_ADMIN")
	globalConfig.PB.AdminPassword = os.Getenv("PB_ADMIN_PASS")
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
