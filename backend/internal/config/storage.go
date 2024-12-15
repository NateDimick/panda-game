package config

type SurrealConfig struct {
	Address       string
	AdminIdentity string
	AdminPassword string
	Namespace     string
	Database      string
}

type NatsConfig struct {
	Address      string
	RelaySubject string
	GroupBucket  string
}
