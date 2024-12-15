package web

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

const PandaGameCookie = "PGToken"

func GetToken(req *http.Request) (string, error) {
	tokenCookie, _ := req.Cookie(PandaGameCookie)
	tokenHeader := req.Header.Get("Authorization")
	token := ""
	if tokenCookie != nil {
		token = tokenCookie.Value
	}
	if tokenHeader != "" {
		token = tokenHeader
	}
	if token == "" {
		return "", errors.New("no token")
	}
	return token, nil
}

func IDFromRequest(req *http.Request) string {
	token, _ := GetToken(req)
	return IDFromToken(token)
}

func IDFromToken(token string) string {
	slog.Info("Cracking open token", slog.String("token", token))
	if token == "" {
		return token
	}
	middle := strings.Split(token, ".")[1]
	raw, err := base64.URLEncoding.DecodeString(middle)
	if err != nil {
		return ""
	}
	claims := make(map[string]any)
	if err := json.NewDecoder(bytes.NewReader(raw)).Decode(&claims); err != nil {
		return ""
	}
	return claims["ID"].(string)
}
