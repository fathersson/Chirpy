package config

import "sync/atomic"

type Config struct {
	DbUrl          string `env:"DbUrl"`
	Platform       string `env:"Platform"`
	TokenSecret    string `env:"TokenSecret"`
	PolkaKey       string `env:"PolkaKey"`
	FileserverHits atomic.Int32
}

func NewConfig() *Config {
	return &Config{}
}
