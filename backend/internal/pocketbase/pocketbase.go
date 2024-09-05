package pocketbase

import (
	"encoding/json"
	"net/http"
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
		p.holder = &tokenHolder{config: p.config}
	}
	if p.mode != adminUser {
		p.holder.token = ""
	}
	p.mode = adminUser
	return p.holder
}

func (p *PBClient) AsUser() API {
	if p.holder == nil {
		p.holder = &tokenHolder{config: p.config}
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
	return p
}

type AdminAPI interface {
	Collections(string) CollectionsAPI
	Admins() AdminsAPI
	Records(string) RecordsAPI
	Realtime() RealtimeAPI
	// Backups
	// Settings
	// Logs
}

type API interface {
	Records(string) RecordsAPI
	Auth(string) AuthAPI
	Realtime() RealtimeAPI
	// Files
	// Health
}

type tokenHolder struct {
	config *pbConfig
	token  string
}

func (t *tokenHolder) Admins() AdminsAPI {
	return t
}

func (t *tokenHolder) Realtime() RealtimeAPI {
	return t
}

func (t *tokenHolder) Collections(collection string) CollectionsAPI {
	return &collectionsClient{
		tokenHolder: t,
		collection:  collection,
	}
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

type noResponse struct{}

func (*noResponse) UnmarshalJSON([]byte) error {
	return nil
}
