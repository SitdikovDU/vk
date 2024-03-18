package config

import "github.com/ilyakaznacheev/cleanenv"

type Database struct {
	DriverName     string `env:"DB_DRIVER" env-default:"postgres"`
	DataSourceName string `env:"DB_DSN"`
}

func NewDatabase() (*Database, error) {
	var cfg Database
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
