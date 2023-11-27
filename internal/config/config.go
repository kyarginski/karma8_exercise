package config

import (
	"log"
	"os"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env       string `yaml:"env" env-default:"local"`
	Version   string `yaml:"version" env-default:"unknown"`
	Port      int    `yaml:"port" env-default:"8085"`
	DBConnect string `yaml:"db_connect" env-default:""`
}

func MustLoad(name string) *Config {
	configPath := os.Getenv(strings.ToUpper(name) + "_CONFIG_PATH")
	if configPath == "" {
		configPath = "config/" + name + "/local.yaml"
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
