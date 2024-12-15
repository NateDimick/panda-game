package web

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDFromToken(t *testing.T) {
	token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzUxMiJ9.eyJpYXQiOjE3MzQyODQ1MTIsIm5iZiI6MTczNDI4NDUxMiwiZXhwIjoxNzQzOTYxMzEyLCJpc3MiOiJTdXJyZWFsREIiLCJqdGkiOiI3ZmEyNzdiMS0xODFjLTRkYTgtYTU2OC00OWU1ZmM0MjJjOGMiLCJOUyI6InBhbmRhTlMiLCJEQiI6InBhbmRhREIiLCJBQyI6InBsYXllciIsIklEIjoicGxheWVyOm5hdGUifQ.-tlJP0Rcrl0uyKWLyb3ZZyOkD1q2gcxcvInJwT6V1sAnzZTRJOZ7ux4E6LA8FPiEfxj1ct6zkM5rGznImBtZFQ"
	id := IDFromToken(token)
	assert.NotEmpty(t, id)
}
