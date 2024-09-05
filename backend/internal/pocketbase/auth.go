package pocketbase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type AuthAPI interface {
	Create(NewAuthRecord) (Record, error)
	PasswordAuth(AuthPasswordBody) (AuthResponse, error)
	RefreshAuth(RecordQuery) (AuthResponse, error)
}

type authClient struct {
	*tokenHolder
	collection string
}

type NewAuthRecord struct {
	Username        string
	Password        string
	ConfirmPassword string
}

type AuthPasswordBody struct {
	Username string `json:"identity"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token  string         `json:"token"`
	Record map[string]any `json:"record"`
}

// https://pocketbase.io/docs/api-records/#create-record
func (a *authClient) Create(record NewAuthRecord) (Record, error) {

	plainRecord := NewRecord{
		Fields: map[string]any{
			"username":        record.Username,
			"password":        record.Password,
			"confirmPassword": record.ConfirmPassword,
		},
	}
	return a.Records(a.collection).Create(plainRecord, RecordQuery{})
}

// https://pocketbase.io/docs/api-records/#auth-with-password
func (a *authClient) PasswordAuth(credentials AuthPasswordBody) (AuthResponse, error) {
	body := bytes.NewBuffer(make([]byte, 0))
	if err := json.NewEncoder(body).Encode(credentials); err != nil {
		return AuthResponse{}, err
	}
	url := fmt.Sprintf("%s/api/collections/%s/auth-with-password", a.config.Addr, a.collection)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return AuthResponse{}, err
	}
	req.Header.Add("Content-Type", "application/json")
	auth, err := handleResponse[AuthResponse](a.config.Client.Do(req))
	a.setToken(auth)
	return auth, err
}

// https://pocketbase.io/docs/api-records/#auth-refresh
func (a *authClient) RefreshAuth(query RecordQuery) (AuthResponse, error) {
	url := fmt.Sprintf("%s/api/collections/%s/auth-refresh", a.config.Addr, a.collection)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return AuthResponse{}, err
	}
	req.Header.Add("Authorization", a.token)
	auth, err := handleResponse[AuthResponse](a.config.Client.Do(req))
	a.setToken(auth)
	return auth, err
}

func (a *authClient) setToken(auth AuthResponse) {
	if auth.Token != "" {
		a.token = auth.Token
	}
}
