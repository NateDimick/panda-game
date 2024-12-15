package sign

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"pandagame/internal/config"
)

// surreal go sdk 0.3.2 and surrealdb 2.1.x don't work together for signin and signup - but the http endpoint does work.

type SurrealTokenResponse struct {
	Code    int    `json:"code"`
	Details string `json:"details"`
	Token   string `json:"token"`
}

type SurrealHTTPError struct {
	Code        int    `json:"code"`
	Details     string `json:"details"`
	Description string `json:"description"`
	Info        string `json:"information"`
}

func (e *SurrealHTTPError) Error() string {
	return fmt.Sprintf("%#v", *e)
}

func Up(username, password string) error {
	cfg := config.LoadAppConfig()
	req, err := http.NewRequest(http.MethodPost, cfg.Surreal.HTTPAddress+"/signup", prepareBody(username, password))
	req.Header.Add("Accept", "application/json")
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		err, derr := unmarshalResp[SurrealHTTPError](resp)
		return errors.Join(err, derr)
	}
	return nil
}

func In(username, password string) (string, error) {
	cfg := config.LoadAppConfig()
	req, err := http.NewRequest(http.MethodPost, cfg.Surreal.HTTPAddress+"/signin", prepareBody(username, password))
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	} else if resp.StatusCode != http.StatusOK {
		err, derr := unmarshalResp[SurrealHTTPError](resp)
		return "", errors.Join(err, derr)
	}
	body, err := unmarshalResp[SurrealTokenResponse](resp)
	if err != nil {
		return "", err
	}
	return body.Token, nil
}

func prepareBody(username, password string) io.Reader {
	cfg := config.LoadAppConfig()
	body := map[string]string{
		"AC":       "player",
		"DB":       cfg.Surreal.Database,
		"NS":       cfg.Surreal.Namespace,
		"name":     username,
		"password": password,
	}
	bb := bytes.NewBuffer(make([]byte, 0))
	json.NewEncoder(bb).Encode(&body)
	return bb
}

func unmarshalResp[T any](resp *http.Response) (*T, error) {
	thing := new(T)
	if err := json.NewDecoder(resp.Body).Decode(thing); err != nil {
		return nil, err
	}
	resp.Body.Close()
	return thing, nil
}
