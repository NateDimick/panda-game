package config

import (
	"bufio"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetLogger(t *testing.T) {
	SetLogger("")
	slog.Info("unit test message", slog.String("extraValue", "hello"))
	f, err := os.Open("panda-server.log")
	assert.Nil(t, err)

	msg := make(map[string]interface{})
	json.NewDecoder(f).Decode(&msg)
	f.Close()

	assert.Equal(t, "unit test message", msg["msg"])
	t.Cleanup(func() {
		os.Remove("panda-server.log")
	})
}

func TestTeeHandler(t *testing.T) {
	SetLogger("")
	slog.Default().WithGroup("groupOne").Info("the bees!", slog.String("valueOne", "bees"))
	slog.With(slog.String("attrOne", "knees")).Info("the knees!")

	f, err := os.Open("panda-server.log")
	assert.Nil(t, err)

	scanner := bufio.NewScanner(f)
	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	msg1 := make(map[string]interface{})
	msg2 := make(map[string]interface{})
	json.NewDecoder(strings.NewReader(lines[0])).Decode(&msg1)
	json.NewDecoder(strings.NewReader(lines[1])).Decode(&msg2)

	assert.NotNil(t, msg1["groupOne"])
	assert.Equal(t, "knees", msg2["attrOne"])
	t.Cleanup(func() {
		os.Remove("panda-server.log")
	})
}
