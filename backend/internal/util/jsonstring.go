package util

import "encoding/json"

func FromJSONString[T any](s string) (*T, error) {
	t := new(T)
	if err := json.Unmarshal([]byte(s), t); err != nil {
		return nil, err
	}
	return t, nil
}

func ToJSONString[T any](v *T) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
