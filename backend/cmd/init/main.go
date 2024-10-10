package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"pandagame/internal/config"
	"pandagame/internal/pocketbase"
)

func pBool(b bool) *bool {
	return &b
}

func pInt(i int) *int {
	return &i
}

func main() {
	config.SetLogger("")
	addr := os.Getenv("PB_ADDR")
	pb := pocketbase.NewPocketBase(addr, nil)
	creds := pocketbase.AdminPasswordBody{
		Identity: os.Getenv("PB_ADMIN"),
		Password: os.Getenv("PB_ADMIN_PASS"),
	}
	initCreds := pocketbase.NewAdminBody{
		Email:    creds.Identity,
		Password: creds.Password,
	}
	pbAdmin := pb.AsAdmin()
	pbe := new(pocketbase.PocketbaseError)
	if err := pbAdmin.Admins().CreateAdmin(initCreds); err != nil {
		if errors.As(err, &pbe) {
			if pbe.Code != http.StatusUnauthorized {
				slog.Error(pbe.Error())
				os.Exit(1)
			} else {
				slog.Info("admin already exists")
			}
		} else {
			slog.Error(pbe.Error())
			os.Exit(1)
		}
	}

	if _, err := pbAdmin.Admins().PasswordAuth(creds); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	slog.Info("logged in")

	// need to create user collection, game collection, lobby collection (?)
	slog.Info("creating players collection")
	resp, err := pbAdmin.Collections().Create(pocketbase.NewAuthCollection{
		NewBaseCollection: pocketbase.NewBaseCollection{
			Collection: pocketbase.Collection{
				Name: "players",
				Type: "auth",
			},
			System: true,
		},
		RequireEmail:      pBool(false),
		MinPasswordLength: pInt(12),
		AllowUsernameAuth: pBool(true),
	}, pocketbase.CollectionQuery{})

	if err != nil {
		slog.Error(err.Error())
		if errors.As(err, &pbe) {
			if pbe.Code != 400 {
				os.Exit(1)
			}
		}
	} else {
		slog.Info(fmt.Sprintf("created player collection: %+v", resp))
	}

	slog.Info("creating games collection")
	resp, err = pbAdmin.Collections().Create(pocketbase.NewBaseCollection{
		Collection: pocketbase.Collection{
			Name: "games",
			Type: "base",
			Schema: []pocketbase.SchemaItem{
				{
					Name: "gameId",
					Type: "text",
				},
				{
					Name: "state",
					Type: "json",
					Options: map[string]any{
						"maxSize": 20000,
					},
				},
			},
		},
		System: *pBool(true),
	}, pocketbase.CollectionQuery{})

	if err != nil {
		slog.Error(err.Error())
		if errors.As(err, &pbe) {
			if pbe.Code != 400 {
				os.Exit(1)
			}
		}
	} else {
		slog.Info(fmt.Sprintf("created game collection: %+v", resp))
	}

	slog.Info("done")
}
