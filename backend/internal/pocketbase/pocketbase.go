package pocketbase

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type userType int

const (
	adminUser userType = iota
	normalUser
)

type PBClient struct {
	config *pbConfig
	holder *tokenHolder
	mode   userType
}

type pbConfig struct {
	Addr   string
	Client *http.Client
}

// start here to begin using pocketbase
func NewPocketBase(addr string, client *http.Client) *PBClient {
	if client == nil {
		client = http.DefaultClient
	}
	return &PBClient{
		config: &pbConfig{
			Addr:   addr,
			Client: client,
		},
	}
}

func (p *PBClient) AsAdmin() AdminAPI {
	if p.holder == nil {
		p.holder = &tokenHolder{config: p.config, refresher: &tokenRefresher{mode: adminUser}}
	}
	if p.mode != adminUser {
		p.holder.token = ""
	}
	p.mode = adminUser
	return p.holder
}

func (p *PBClient) AsUser() API {
	if p.holder == nil {
		p.holder = &tokenHolder{config: p.config, refresher: &tokenRefresher{mode: normalUser}}
	}
	if p.mode != normalUser {
		p.holder.token = ""
	}
	p.mode = normalUser
	return p.holder
}

func (p *PBClient) WithToken(token string) *PBClient {
	p.AsUser()
	p.holder.token = token
	p.holder.refresher.refreshTime = getExpiryTime(token)
	p.holder.refresher.collection = getCollectionID(token)
	return p
}

type AdminAPI interface {
	Collections() CollectionsAPI
	Admins() AdminsAPI
	Records(string) RecordsAPI
	AdminAuth(string) AdminAuthAPI
	// Backups
	// Settings
	// Logs
}

type API interface {
	Records(string) RecordsAPI
	Auth(string) AuthAPI
	// Files
	// Health
}

type tokenHolder struct {
	config    *pbConfig
	token     string
	refresher *tokenRefresher
}

func (t *tokenHolder) Admins() AdminsAPI {
	return t
}

func (t *tokenHolder) Collections() CollectionsAPI {
	return t
}

func (t *tokenHolder) Records(collection string) RecordsAPI {
	return &recordClient{
		collection:  collection,
		tokenHolder: t,
	}
}

func (t *tokenHolder) Auth(collection string) AuthAPI {
	return &authClient{
		collection:  collection,
		tokenHolder: t,
	}
}

func (t *tokenHolder) AdminAuth(collection string) AdminAuthAPI {
	return &authClient{
		collection:  collection,
		tokenHolder: t,
	}
}

func prepareRequest(method, url string, payload any, t *tokenHolder) (*http.Request, error) {
	bb := bytes.NewBuffer(make([]byte, 0))
	if payload != nil {
		if err := json.NewEncoder(bb).Encode(payload); err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, url, bb)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	if t.token != "" {
		t.refresher.refreshToken(t)
		req.Header.Add("Authorization", t.token)
	}
	slog.Info("prepared pocketbase request", slog.String("body", bb.String()), slog.Any("headers", req.Header), slog.String("method", method), slog.String("url", url))
	return req, nil
}

func handleResponse[T any](resp *http.Response, err error) (T, error) {
	defer resp.Body.Close()
	empty := *new(T)
	if err != nil {
		return empty, err
	}
	if err := getPocketbaseError(resp); err != nil {
		return empty, err
	}
	data := new(T)
	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return empty, err
	}
	return *data, err
}

type tokenRefresher struct {
	username    string
	password    string
	mode        userType
	collection  string
	refreshTime time.Time
}

func (t *tokenRefresher) refreshToken(h *tokenHolder) {
	if time.Now().Before(t.refreshTime) {
		return
	}
	switch t.mode {
	case adminUser:
		//
		if _, err := h.Admins().RefreshAuth(); err != nil {
			h.Admins().PasswordAuth(AdminPasswordBody{t.username, t.password})
		}
	case normalUser:
		fallthrough
	default:
		//
		if _, err := h.Auth(t.collection).RefreshAuth(nil); err != nil {
			h.Auth(t.collection).PasswordAuth(AuthPasswordBody{t.username, t.password})
		}
	}
}
