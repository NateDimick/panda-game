package config

import (
	"errors"
	"net/http"
	"pandagame/internal/pocketbase"
)

type PocketBaseConfig struct {
	Address       string
	AdminIdentity string
	AdminPassword string
}

type NatsConfig struct {
	Address      string
	RelaySubject string
	GroupBucket  string
}

// creates a PB admin. only works the first time. handles failure mode if admin already exists.
func CreatePBAdmin(cfg PocketBaseConfig) error {
	pb := pocketbase.NewPocketBase(cfg.Address, nil)
	err := pb.AsAdmin().Admins().CreateAdmin(pocketbase.AdminPasswordBody{
		Identity:        cfg.AdminIdentity,
		Password:        cfg.AdminPassword,
		ConfirmPassword: cfg.AdminPassword,
	})
	if err != nil {
		pbe := new(pocketbase.PocketbaseError)
		if errors.As(err, &pbe) {
			if pbe.Code == http.StatusUnauthorized {
				// 401 signals admin already exists. not a real error.
				return nil
			}
		}
	}
	return err
}
