package pocketbase

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type AuthAPI interface {
	Create(NewAuthRecord, any) (Record, error)
	PasswordAuth(AuthPasswordBody) (AuthResponse, error)
	RefreshAuth(*RecordQuery) (AuthResponse, error)
}

type authClient struct {
	*tokenHolder
	collection string
}

type NewAuthCredentials struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type NewAuthRecord struct {
	Credentials  NewAuthCredentials
	CustomFields any
}

func (a *NewAuthRecord) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(a.Credentials)
	if err != nil {
		return nil, err
	}
	if a.CustomFields == nil {
		return b1, nil
	}
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(b1[:len(b1)-1])
	buf.WriteByte(',')
	b2, err := json.Marshal(a.CustomFields)
	if err != nil {
		return nil, err
	}
	buf.Write(b2[1:])
	return buf.Bytes(), nil
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
func (a *authClient) Create(record NewAuthRecord, out any) (Record, error) {
	plainRecord := NewRecord{
		CustomFields: record,
	}
	return a.Records(a.collection).Create(plainRecord, out, nil)
}

// https://pocketbase.io/docs/api-records/#auth-with-password
func (a *authClient) PasswordAuth(credentials AuthPasswordBody) (AuthResponse, error) {
	a.token = ""
	url := fmt.Sprintf("%s/api/collections/%s/auth-with-password", a.config.Addr, a.collection)
	req, err := prepareRequest(http.MethodPost, url, credentials, a.tokenHolder)
	if err != nil {
		return AuthResponse{}, err
	}
	auth, err := handleResponse[AuthResponse](a.config.Client.Do(req))
	a.setToken(auth)
	a.refresher.collection = a.collection
	a.refresher.username = credentials.Username
	a.refresher.password = credentials.Password
	a.refresher.refreshTime = getExpiryTime(a.token)
	return auth, err
}

// https://pocketbase.io/docs/api-records/#auth-refresh
func (a *authClient) RefreshAuth(query *RecordQuery) (AuthResponse, error) {
	url := fmt.Sprintf("%s/api/collections/%s/auth-refresh", a.config.Addr, a.collection)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return AuthResponse{}, err
	}
	req.Header.Add("Authorization", a.token)
	if query != nil {
		req.URL.RawQuery = query.ToQuery()
	}
	auth, err := handleResponse[AuthResponse](a.config.Client.Do(req))
	a.setToken(auth)
	a.refresher.refreshTime = getExpiryTime(a.token)
	return auth, err
}

func (a *authClient) setToken(auth AuthResponse) {
	if auth.Token != "" {
		a.token = auth.Token
	}
}

type pocketbaseJWT struct {
	ID         string `json:"id"`
	Exp        int    `json:"exp"`
	Type       string `json:"type"`
	Collection string `json:"collectionId"`
}

func getExpiryTime(token string) time.Time {
	claims := getClaims(token)
	return time.Unix(int64(claims.Exp), 0)
}

func getCollectionID(token string) string {
	return getClaims(token).Collection
}

func getClaims(token string) *pocketbaseJWT {
	// example admin pb token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mjk3MzY0NDMsImlkIjoibzhsdHFlM2tjd25lOTd6IiwidHlwZSI6ImFkbWluIn0.vyRB-bz7cb2l1ha-1A34p_gOX_jIVOVqjAGxtExODm8
	// {"exp":1729736443,"id":"o8ltqe3kcwne97z","type":"admin"}
	// example user pb token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2xsZWN0aW9uSWQiOiJfcGJfdXNlcnNfYXV0aF8iLCJleHAiOjE3Mjk5MTExNDUsImlkIjoidWR4ODJyNjV6Zm12MjJwIiwidHlwZSI6ImF1dGhSZWNvcmQifQ.PM4Z5Ai92dr_NGsNqbiQMMrmeAolx9O1y5B-LdkzxsM
	// {"collectionId":"_pb_users_auth_","exp":1729911145,"id":"udx82r65zfmv22p","type":"authRecord"}
	jwt := strings.Split(token, ".")
	b64Claims := jwt[1]
	decodedClaims, err := base64.RawURLEncoding.DecodeString(b64Claims)
	if err != nil {
		return nil
	}
	claims := new(pocketbaseJWT)
	json.Unmarshal(decodedClaims, claims)
	return claims
}
