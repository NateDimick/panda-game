package events

import "math/rand"

// 34^6 possible outcomes (>1.5 billion)
func NewGameID() string {
	chars := []rune("ABCDEFGHIJKMNPQRSTUBWXYZ1234567890") // L and ) removed because if L became lowercase it can be hard to tell from 1. same with O and 0.
	id := [6]rune{}
	for i := 0; i < 6; i++ {
		id[i] = chars[rand.Intn(34)]
	}
	return string(id[:])
}
