package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGameID(t *testing.T) {
	id := NewGameID()
	assert.Equal(t, 6, len(id))
}
