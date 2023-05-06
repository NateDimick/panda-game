package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildLogger(t *testing.T) {
	cfg := LoggerConfig()
	logger, err := cfg.Build()
	assert.Nil(t, err)
	assert.NotNil(t, logger)
}
