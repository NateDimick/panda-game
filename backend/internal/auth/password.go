package auth

import (
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"io"
	"math"
	"math/big"
)

// password strategy
// * use username + a server secret to generate the salt
// * interleave the salt and password
// * store salted password hash

//go:embed salt.secret
var secret []byte

type saltReader struct {
	b   []byte
	pos int
}

func (r *saltReader) Read(p []byte) (int, error) {
	l := len(p)
	for i := 0; i < l; i++ {
		p[i] = r.b[r.pos]
		r.pos = (r.pos + 1) % len(r.b)
	}
	return l, nil
}

func newSaltReader(seed []byte) io.Reader {
	return &saltReader{append(secret, seed...), 0}
}

func generateSalt(username []byte, l int) []byte {
	sr := newSaltReader(username)
	s := make([]byte, l)
	for i := 0; i < l; i++ {
		n, _ := rand.Int(sr, big.NewInt(math.MaxInt8))
		s = append(s, byte(n.Int64()))
	}
	return s
}

func interleaveSalt(password, salt []byte) []byte {
	s := make([]byte, len(password)*2)
	for i := 0; i < len(password); i++ {
		s[i*2] = password[i]
		s[i*2+1] = salt[i]
	}
	return s
}

func hashPassword(saltedPass []byte) []byte {
	hasher := sha256.New()
	return hasher.Sum(saltedPass)
}

func NewSalt(username, password []byte) (string, string) {
	salt := generateSalt(username, len(password))
	salted := interleaveSalt(password, salt)
	return hex.EncodeToString(hashPassword(salted)), hex.EncodeToString(salt)
}

func Verify(provided, salt []byte, actual string) bool {
	salted := interleaveSalt(provided, salt)
	return actual == hex.EncodeToString(hashPassword(salted))
}
