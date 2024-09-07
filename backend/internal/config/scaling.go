package config

type ScalingLevel int

const (
	Singleton   ScalingLevel = iota // just one instance. framework can operate in memory
	Colocated                       // multiple instances in the same cluster. redis rooms and pocketbase relay
	Distributed                     // unsupported. would require a different db solution than pocketbase
)
