package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaltPassword(t *testing.T) {
	uname := []byte("edward")
	pass := []byte("helloworld")
	result, _ := NewSalt(uname, pass)
	assert.NotEmpty(t, result)
	// same pair should always produce the same result
	result2, _ := NewSalt(uname, pass)
	assert.Equal(t, result, result2)
}
