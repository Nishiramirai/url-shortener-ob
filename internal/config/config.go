package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `env:"ENV" env-default:"local"`
	StorageType string `env:"STORAGE_TYPE" env-required:"true"`
	HTTPServer  HTTPServerConfig
	Postgres    PostgresConfig
}

type HTTPServerConfig struct {
	Address     string        `env:"HTTP_ADDRESS" env-default:":8080"`
	Timeout     time.Duration `env:"HTTP_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
}

type PostgresConfig struct {
	DSN string `env:"DATABASE_DSN"`
}

func MustLoad() *Config {
	var cfg Config

	if _, err := os.Stat(".env"); err == nil {
		if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
			log.Fatalf("cannot read .env file: %s", err)
		}
	} else {
		log.Println(".env file not found, loading options from system env only")
	}

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		log.Fatalf("cannot read env variables: %s", err)
	}

	return &cfg
}
