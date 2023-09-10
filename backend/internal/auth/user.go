package auth

import "time"

type UserRecord struct {
	Name              string
	PrimaryIdentity   string    // salted password hash
	SecondaryIdentity string    // salt
	TertiaryIdentity  string    // PlayerID
	Guest             bool      // guests don't need passwords, and expire
	Empowered         bool      // empowered users can create games
	ExpireAt          time.Time `bson:"omitempty"` // configure mongo to auto-delete based on this field - only present for guests
}

type UserSession struct {
	SessionID string    // session id
	Name      string    // user name
	PlayerID  string    // player id
	Empowered bool      // if the user is an empowered user
	ExpireAt  time.Time // when the session is over - configure mongo to auto-delete on this field
}
