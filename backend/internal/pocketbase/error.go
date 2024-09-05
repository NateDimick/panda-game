package pocketbase

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type PocketbaseError struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}

func (p *PocketbaseError) Error() string {
	return fmt.Sprintf("Pocketbase Error: %d - %s | %+v", p.Code, p.Message, p.Data)
}

func getPocketbaseError(resp *http.Response) error {
	if resp.StatusCode < 300 {
		return nil
	}
	pbe := new(PocketbaseError)
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(pbe); err != nil {
		return err
	}
	return pbe
}
