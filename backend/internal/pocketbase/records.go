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
	View(string, any, *RecordQuery) (Record, error)
	Create(NewRecord, any, *RecordQuery) (Record, error)
	Update(string, any, any, *RecordQuery) (Record, error)
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
	ID           string `json:"id,omitempty"`
	CustomFields any    // must be a json-marshalable value to an object, such as a struct or map
}

func (r NewRecord) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(r.CustomFields)
	if err != nil {
		return nil, err
	}
	if r.ID == "" {
		return b, nil
	}
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.WriteString(fmt.Sprintf("{\"id\":\"%s\",", r.ID))
	buf.Write(b[1:])
	return buf.Bytes(), nil
}

type Record struct {
	ID             string         `json:"id"`
	CollectionId   string         `json:"collectionId"`
	CollectionName string         `json:"collectionName"`
	UpdatedDT      string         `json:"updated"`
	CreatedDT      string         `json:"created"`
	CustomFields   any            // provide a pointer to a type that matches custom fields. if nil, will result in map[string]any
	Expand         map[string]any `json:"expand"`
}

type aliasRecord Record

type unmarshalingRecord struct {
	*aliasRecord
	AltCollID   string `json:"@collectionId"`
	AltCollName string `json:"@collectionName"`
}

func (r *Record) UnmarshalJSON(b []byte) error {
	u := &unmarshalingRecord{aliasRecord: (*aliasRecord)(r)}
	err := json.Unmarshal(b, u)
	if err != nil {
		return err
	}
	if r.CustomFields != nil {
		err := json.Unmarshal(b, r.CustomFields)
		if err != nil {
			return err
		}
		if u.AltCollID != "" {
			u.CollectionId = u.AltCollID
		}
		if u.AltCollName != "" {
			r.CollectionName = u.AltCollName
		}
	} else {
		m := make(map[string]any)
		err := json.Unmarshal(b, &m)
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
	}

	return nil
}

type recordClient struct {
	*tokenHolder
	collection string
}

// https://pocketbase.io/docs/api-records/#view-record
func (r *recordClient) View(recordId string, out any, query *RecordQuery) (Record, error) {
	url := fmt.Sprintf("%s/api/collections/%s/records/%s", r.config.Addr, r.collection, recordId)
	req, err := prepareRequest(http.MethodGet, url, nil, r.tokenHolder)
	if err != nil {
		return Record{}, nil
	}
	if query != nil {
		req.URL.RawQuery = query.ToQuery()
	}
	resp, err := r.config.Client.Do(req)
	return handleRecordResponse(out, resp, err)
}

// https://pocketbase.io/docs/api-records/#create-record
func (r *recordClient) Create(record NewRecord, out any, query *RecordQuery) (Record, error) {
	url := fmt.Sprintf("%s/api/collections/%s/records", r.config.Addr, r.collection)
	req, _ := prepareRequest(http.MethodPost, url, record, r.tokenHolder)
	if query != nil {
		req.URL.RawQuery = query.ToQuery()
	}
	resp, err := r.config.Client.Do(req)
	return handleRecordResponse(out, resp, err)
}

// https://pocketbase.io/docs/api-records/#update-record
func (r *recordClient) Update(recordId string, update, out any, query *RecordQuery) (Record, error) {
	url := fmt.Sprintf("%s/api/collections/%s/records/%s", r.config.Addr, r.collection, recordId)
	req, _ := prepareRequest(http.MethodPatch, url, update, r.tokenHolder)
	if query != nil {
		req.URL.RawQuery = query.ToQuery()
	}
	resp, err := r.config.Client.Do(req)
	return handleRecordResponse(out, resp, err)
}

// https://pocketbase.io/docs/api-records/#delete-record
func (r *recordClient) Delete(recordId string) error {
	url := fmt.Sprintf("%s/api/collections/%s/records/%s", r.config.Addr, r.collection, recordId)
	req, _ := prepareRequest(http.MethodDelete, url, nil, r.tokenHolder)
	_, err := handleResponse[Record](r.config.Client.Do(req))
	return err
}

func handleRecordResponse(out any, resp *http.Response, err error) (Record, error) {
	defer resp.Body.Close()
	empty := *new(Record)
	if err != nil {
		return empty, err
	}
	if err := getPocketbaseError(resp); err != nil {
		return empty, err
	}
	data := new(Record)
	if out != nil {
		data.CustomFields = out
	}
	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return empty, err
	}
	return *data, nil
}
