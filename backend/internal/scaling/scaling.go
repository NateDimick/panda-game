package scaling

import (
	"pandagame/internal/config"
	"pandagame/internal/framework"
)

func Rooms(cfg config.AppConfig) framework.Rooms {
	switch cfg.Scale {
	case config.Singleton:
		return framework.NewInMemStorage()
	case config.Colocated:
		return NewRedisRooms(cfg.Redis)
	case config.Distributed:
		panic("distributed is not possible yet")
	default:
		panic("invalid scaling level")
	}
}

func Relayer(cfg config.AppConfig) framework.Relayer {
	switch cfg.Scale {
	case config.Singleton:
		return framework.NewInMemRelayer()
	case config.Colocated:
		return NewPocketBaseRelay(cfg.PB)
	case config.Distributed:
		panic("distributed is not possible yet")
	default:
		panic("invalid scaling level")
	}
}
