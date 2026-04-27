package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env:"ENV" env-default:"local"` //struct tag
	StoragePath string `yaml:"storage_path" env:"SPath" env-required:"true"`
	DriverName  string `yaml:"driver_name" env:"Drv" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
}

type HTTPServer struct {
	Address     string        `yml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	//panic("not implement")

	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatal("config file does not exist:", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatal("cannot read config:", err)
	}

	return &cfg
}
