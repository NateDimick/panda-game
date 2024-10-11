package pocketbase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaBuilder(t *testing.T) {
	s1 := SchemaBuilder("jsonexample", JSONOptions{MaxSize: 37})
	sc := s1.(schema)
	assert.Equal(t, sc.Type, "json")

	s2 := SchemaBuilder("fileexample", FileOptions{})
	sc = s2.(schema)
	assert.Equal(t, sc.Type, "file")
}
