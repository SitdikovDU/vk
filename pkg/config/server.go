package config

import "github.com/ilyakaznacheev/cleanenv"

type Server struct {
	Addr string `env:"SERVER_ADDR" env-default:"0.0.0.0:8080"`
}

func NewServer() (*Server, error) {
	var cfg Server
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
