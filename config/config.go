package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTP struct {
		IP   string `env-required:"true" yaml:"ip" env:"APP_IP"`
		Port string `env-required:"true" yaml:"port" env:"APP_PORT"`
	}
	Postgres struct {
		Password string `env-required:"true" yaml:"password" env:"DB_PASSWORD"`
		Username string `env-required:"true" yaml:"username" env:"DB_USERNAME"`
		Host     string `env-required:"true" yaml:"host" env:"DB_HOST"`
		Port     string `env-required:"true" yaml:"port" env:"DB_PORT"`
		Database string `env-required:"true" yaml:"database" env:"DB_DATABASE"`
	} `yaml:"postgresql"`
}

func GetConfig() *Config {

	log.Print("config init")

	c := &Config{}

	if err := cleanenv.ReadConfig("config/config.yaml", c); err != nil {

	}

	return c
}