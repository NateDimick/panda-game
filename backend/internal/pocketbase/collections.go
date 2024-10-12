package pocketbase

import (
	"errors"
	"fmt"
	"net/http"
)

type CollectionsAPI interface {
	View(string, CollectionQuery) (CollectionResponse, error)
	Create(NewCollection, CollectionQuery) (CollectionResponse, error)
}

type CollectionQuery struct {
	Fields []string
}

type NewCollection interface {
	C()
}

type Schema interface {
	S()
}

type Collection struct {
	ID         string   `json:"id,omitempty"`
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Schema     []Schema `json:"schema"`
	ListRule   string   `json:"listRule,omitempty"`
	ViewRule   string   `json:"viewRule,omitempty"`
	CreateRule string   `json:"createRule,omitempty"`
	UpdateRule string   `json:"updateRule,omitempty"`
	DeleteRule string   `json:"deleteRule,omitempty"`
	Indexes    []string `json:"indexes,omitempty"`
}

func (c Collection) C() {}

type NewBaseCollection struct {
	Collection
	System bool `json:"system,omitempty"`
}

type NewViewCollection struct {
	NewBaseCollection
	Query string `json:"query,omitempty"`
}

type NewAuthCollection struct {
	NewBaseCollection
	ManageRule         string   `json:"manageRule,omitempty"`
	AllowOAuth2Auth    *bool    `json:"allowOAuth2Auth,omitempty"`
	AllowUsernameAuth  *bool    `json:"allowUsernameAuth,omitempty"`
	AllowEmailAuth     *bool    `json:"allowEmailAuth,omitempty"`
	RequireEmail       *bool    `json:"requireEmail,omitempty"`
	ExceptEmailDomains []string `json:"exceptEmailDomains,omitempty"`
	OnlyEmailDomains   []string `json:"onlyEmailDomains,omitempty"`
	OnlyVerified       *bool    `json:"onlyVerified,omitempty"`
	MinPasswordLength  *int     `json:"minPasswordLength,omitempty"`
}

type CollectionResponse struct {
	Collection
	Options map[string]any `json:"options"`
}

// https://pocketbase.io/docs/api-collections/#view-collection
func (t *tokenHolder) View(collection string, query CollectionQuery) (CollectionResponse, error) {
	url := fmt.Sprintf("%s/api/collections/%s", t.config.Addr, collection)
	req, err := prepareRequest(http.MethodGet, url, nil, t)
	if err != nil {
		return CollectionResponse{}, err
	}
	return handleResponse[CollectionResponse](t.config.Client.Do(req))
}

// https://pocketbase.io/docs/api-collections/#create-collection
func (t *tokenHolder) Create(col NewCollection, query CollectionQuery) (CollectionResponse, error) {
	switch col.(type) {
	case NewBaseCollection, NewAuthCollection, NewViewCollection:
	default:
		return CollectionResponse{}, errors.New("NewCollection is not of the right type")
	}
	req, err := prepareRequest(http.MethodPost, fmt.Sprintf("%s/api/collections", t.config.Addr), col, t)
	if err != nil {
		return CollectionResponse{}, err
	}
	return handleResponse[CollectionResponse](t.config.Client.Do(req))
}
