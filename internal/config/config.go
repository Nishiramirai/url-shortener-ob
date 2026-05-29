package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env                string `env:"ENV" env-default:"local"`
	StorageType        string `env:"STORAGE_TYPE" env-required:"true"`
	MemoryStorageLimit int    `env:"MEMORY_STORAGE_LIMIT" env-default:"500000"`
	HTTPServer         HTTPServerConfig
	Postgres           PostgresConfig
}

type HTTPServerConfig struct {
	Port        string        `env:"APP_PORT" env-default:"8080"`
	Timeout     time.Duration `env:"HTTP_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
}

type PostgresConfig struct {
	User     string `env:"DB_USER" env-required:"true"`
	Password string `env:"DB_PASSWORD" env-required:"true"`
	Host     string `env:"DB_HOST" env-default:"localhost"`
	Port     string `env:"DB_PORT" env-default:"5432"`
	Name     string `env:"DB_NAME" env-required:"true"`
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

func (p PostgresConfig) ConnectionURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		p.User, p.Password, p.Host, p.Port, p.Name)
}
