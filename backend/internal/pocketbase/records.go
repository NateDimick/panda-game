package pocketbase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type RecordsAPI interface {
	View(string, RecordQuery) (Record, error)
	Create(NewRecord, RecordQuery) (Record, error)
	Update(string, map[string]any, RecordQuery) (Record, error)
	Delete(string) error
}

type RecordQuery struct {
	Expand []string
	Fields []string // don't use :excerpt() syntax. it will break this implementation
}

func (r RecordQuery) ToQuery() string {
	v := make(url.Values)
	if len(r.Fields) > 0 {
		v.Set("fields", strings.Join(r.Fields, ","))
	}
	if len(r.Expand) > 0 {
		v.Set("expand", strings.Join(r.Expand, ","))
	}

	return v.Encode()
}

type NewRecord struct {
	ID     string
	Fields map[string]any
}

type Record struct {
	ID             string `json:"id"`
	CollectionId   string `json:"collectionId"`
	CollectionName string `json:"collectionName"`
	UpdatedDT      string `json:"updated"`
	CreatedDT      string `json:"created"`
	CustomFields   map[string]json.RawMessage
	Expand         map[string]any `json:"expand"`
}

type aliasRecord Record

type unmarshalingRecord struct {
	*aliasRecord
	AltCollID   string `json:"@collectionId"`
	AltCollName string `json:"@collectionName"`
}

func (r *Record) UnmarshalJSON(b []byte) error {
	m := make(map[string]json.RawMessage)
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	u := &unmarshalingRecord{aliasRecord: (*aliasRecord)(r)}
	err = json.Unmarshal(b, u)
	if err != nil {
		return err
	}
	delete(m, "id")
	delete(m, "updated")
	delete(m, "created")
	if u.AltCollID != "" {
		u.CollectionId = u.AltCollID
		delete(m, "@collectionId")
	} else {
		delete(m, "collectionId")
	}

	if u.AltCollName != "" {
		r.CollectionName = u.AltCollName
		delete(m, "@collectionName")
	} else {
		delete(m, "collectionName")
	}
	r.CustomFields = m
	return nil
}

type recordClient struct {
	*tokenHolder
	collection string
}

// https://pocketbase.io/docs/api-records/#view-record
func (r *recordClient) View(recordId string, query RecordQuery) (Record, error) {
	url := fmt.Sprintf("%s/api/collections/%s/records/%s", r.config.Addr, r.collection, recordId)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("Authorization", r.token)
	req.URL.RawQuery = query.ToQuery()
	return handleResponse[Record](r.config.Client.Do(req))
}

// https://pocketbase.io/docs/api-records/#create-record
func (r *recordClient) Create(record NewRecord, query RecordQuery) (Record, error) {
	if record.ID != "" {
		record.Fields["id"] = record.ID
	}
	body := bytes.NewBuffer(make([]byte, 0))
	if err := json.NewEncoder(body).Encode(record.Fields); err != nil {
		return Record{}, err
	}
	url := fmt.Sprintf("%s/api/collections/%s/records", r.config.Addr, r.collection)
	req, _ := http.NewRequest(http.MethodPost, url, body)
	req.Header.Add("Authorization", r.token)
	req.Header.Add("Content-Type", "application/json")
	req.URL.RawQuery = query.ToQuery()
	return handleResponse[Record](r.config.Client.Do(req))
}

// https://pocketbase.io/docs/api-records/#update-record
func (r *recordClient) Update(recordId string, update map[string]any, query RecordQuery) (Record, error) {
	body := bytes.NewBuffer(make([]byte, 0))
	if err := json.NewEncoder(body).Encode(update); err != nil {
		return Record{}, err
	}
	url := fmt.Sprintf("%s/api/collections/%s/records/%s", r.config.Addr, r.collection, recordId)
	req, _ := http.NewRequest(http.MethodPatch, url, body)
	req.Header.Add("Authorization", r.token)
	req.Header.Add("Content-Type", "application/json")
	req.URL.RawQuery = query.ToQuery()
	return handleResponse[Record](r.config.Client.Do(req))
}

// https://pocketbase.io/docs/api-records/#delete-record
func (r *recordClient) Delete(recordId string) error {
	url := fmt.Sprintf("%s/api/collections/%s/records/%s", r.config.Addr, r.collection, recordId)
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	req.Header.Add("Authorization", r.token)
	_, err := handleResponse[Record](r.config.Client.Do(req))
	return err
}
