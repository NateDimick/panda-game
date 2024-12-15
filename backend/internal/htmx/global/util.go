package global

import (
	"net/http"
	"pandagame/internal/config"
	"pandagame/internal/web"
)

func IsAuthenticatedRequest(r *http.Request) (string, error) {
	token, err := web.GetToken(r)
	if err != nil {
		return "", err
	}
	if token != "" {
		db, _ := config.Surreal()
		if err := db.Authenticate(token); err != nil {
			return "", err
		}
	}
	return token, nil
}
