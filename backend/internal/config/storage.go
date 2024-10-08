package config

type PocketBaseConfig struct {
	Address       string
	AdminIdentity string
	AdminPassword string
}

type NatsConfig struct {
	Address      string
	RelaySubject string
	GroupBucket  string
}
