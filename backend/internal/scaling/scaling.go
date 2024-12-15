package scaling

import (
	"pandagame/internal/config"
	"pandagame/internal/framework"
)

func Grouper(cfg config.AppConfig) framework.Grouper {
	switch cfg.Scale {
	case config.Singleton:
		return framework.NewInMemStorage()
	case config.Colocated:
		return NewNatsKV(cfg.Nats.Address, cfg.Nats.GroupBucket)
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
		return NewNatsRelay(cfg.Nats.Address, cfg.Nats.RelaySubject)
	case config.Distributed:
		panic("distributed is not possible yet")
	default:
		panic("invalid scaling level")
	}
}
