package pocketbase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type AdminsAPI interface {
	PasswordAuth(AdminPasswordBody) (AdminAuthResponse, error)
	RefreshAuth() (AdminAuthResponse, error)
	CreateAdmin(NewAdminBody) error
}

type NewAdminBody struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"passwordConfirm,omitempty"` // only needed for creating admins
}

type AdminPasswordBody struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
}

type AdminAuthResponse struct {
	Token string         `json:"token"`
	Admin map[string]any `json:"admin"`
}

// https://pocketbase.io/docs/api-admins/#auth-with-password
func (t *tokenHolder) PasswordAuth(credentials AdminPasswordBody) (AdminAuthResponse, error) {
	t.token = ""
	url := fmt.Sprintf("%s/api/admins/auth-with-password", t.config.Addr)
	req, err := prepareRequest(http.MethodPost, url, credentials, t)
	if err != nil {
		return AdminAuthResponse{}, err
	}
	req.Header.Add("Content-Type", "application/json")
	auth, err := handleResponse[AdminAuthResponse](t.config.Client.Do(req))
	t.setToken(auth)
	t.refresher.username = credentials.Identity
	t.refresher.password = credentials.Password
	t.refresher.refreshTime = getExpiryTime(t.token)
	return auth, err
}

// https://pocketbase.io/docs/api-admins/#auth-refresh
func (t *tokenHolder) RefreshAuth() (AdminAuthResponse, error) {
	url := fmt.Sprintf("%s/api/admins/auth-refresh", t.config.Addr)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return AdminAuthResponse{}, err
	}
	req.Header.Add("Authorization", t.token)
	auth, err := handleResponse[AdminAuthResponse](t.config.Client.Do(req))
	t.setToken(auth)
	t.refresher.refreshTime = getExpiryTime(t.token)
	return auth, err
}

func (t *tokenHolder) setToken(auth AdminAuthResponse) {
	if auth.Token != "" {
		t.token = auth.Token
	}
}

// https://pocketbase.io/docs/api-admins/#create-admin
func (t *tokenHolder) CreateAdmin(credentials NewAdminBody) error {
	credentials.ConfirmPassword = credentials.Password
	body := bytes.NewBuffer(make([]byte, 0))
	if err := json.NewEncoder(body).Encode(credentials); err != nil {
		return err
	}
	url := fmt.Sprintf("%s/api/admins", t.config.Addr)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	if t.token != "" {
		// auth is optional, first time admin can be created without auth
		req.Header.Add("Authorization", t.token)
	}
	_, err = handleResponse[AdminAuthResponse](t.config.Client.Do(req))
	return err
}
