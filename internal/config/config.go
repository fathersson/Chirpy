package config

import "sync/atomic"

type Config struct {
	DBURL          string `env:"DB_URL"`
	PLATFORM       string `env:"PLATFORM"`
	TOKENSECRET    string `env:"TOKEN_SECRET"`
	POLKAKEY       string `env:"POLKA_KEY"`
	FILESERVERHITS atomic.Int32
}

func NewConfig() *Config {
	return &Config{}
}
