package auth

import "time"

type UserRecord struct {
	Name              string
	PrimaryIdentity   string    // salted password hash
	SecondaryIdentity string    // salt
	TertiaryIdentity  string    // PlayerID
	Guest             bool      // guests don't need passwords, and expire
	Empowered         bool      // empowered users can create games
	ExpireAt          time.Time `bson:"omitempty"`
}
