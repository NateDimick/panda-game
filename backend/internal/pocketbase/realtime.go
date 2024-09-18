package pocketbase

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type RealtimeAPI interface {
	Connect(func(RealtimeEvent)) error
	SetSubscriptions(Subscription) error
}

type RealtimeEvent struct {
	Event    string
	Data     map[string]any
	Error    bool
	Comments []string
}

func ParseEventStream(message string) RealtimeEvent {
	event := RealtimeEvent{}
	lines := strings.Split(message, "\n")
	data := make([]string, 0)
	for _, l := range lines {
		label, message, _ := strings.Cut(l, ": ")
		if label == "event" {
			event.Event = label
		} else if label == "data" {
			data = append(data, message)
		} else if label == "" {
			event.Comments = append(event.Comments, message)
		}
	}
	event.Data = make(map[string]any)
	json.NewDecoder(strings.NewReader(strings.Join(data, "\n"))).Decode(&event.Data)
	return event
}

func ContainsCompleteEvent(data []byte) (int, []byte) {
	lines := strings.Split(string(data), "\n")
	contentLen := 0
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			return contentLen + len(l) + 1, data[:contentLen]
		}
		contentLen += len(l) + 1
	}
	return -1, nil
}

type Subscription struct {
	ClientID      string   `json:"clientId"`
	Subscriptions []string `json:"subscriptions"`
}

// https://pocketbase.io/docs/api-realtime/#connect
func (t *tokenHolder) Connect(handler func(RealtimeEvent)) error {
	url := fmt.Sprintf("%s/api/realtime", t.config.Addr)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := t.config.Client.Do(req)
	if err != nil {
		return err
	}
	// a simplified and specialized adaptation of the r3labs sse client
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 4096), 8192)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		if atEOF {
			return len(data), data, nil
		}

		advance, token = ContainsCompleteEvent(data)
		if advance > 0 {
			return advance, token, nil
		}

		return 0, nil, nil
	})
	go func() {
		defer resp.Body.Close()
		for scanner.Scan() {
			t := scanner.Text()
			if len(strings.TrimSpace(t)) > 0 {
				handler(ParseEventStream(t))
			}
			if err := scanner.Err(); err != nil {
				handler(RealtimeEvent{Event: "Scanner-Error", Data: map[string]any{"error": err.Error()}, Error: true})
				return
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()
	return nil
}

// https://pocketbase.io/docs/api-realtime/#set-subscriptions
func (t *tokenHolder) SetSubscriptions(sub Subscription) error {
	body := bytes.NewBuffer(make([]byte, 0))
	if err := json.NewEncoder(body).Encode(sub); err != nil {
		return err
	}
	url := fmt.Sprintf("%s/api/realtime", t.config.Addr)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", t.token)
	_, err = handleResponse[noResponse](t.config.Client.Do(req))
	return err
}
