package global

import (
	"net/http"
	"pandagame/internal/config"
	"pandagame/internal/pocketbase"
	"pandagame/internal/web"
)

func IsAuthenticatedRequest(r *http.Request) (string, error) {
	token, err := web.GetToken(r)
	if err != nil {
		return "", err
	}
	if token != "" {
		cfg := config.LoadAppConfig()
		_, err := pocketbase.NewPocketBase(cfg.PB.Address, nil).WithToken(token).AsUser().Auth("players").RefreshAuth(nil)
		if err != nil {
			return "", err
		}
	}
	return token, nil
}
