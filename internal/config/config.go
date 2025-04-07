package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server      ServerConfig `yaml:"server"`
	DB          DBConfig     `yaml:"postgres"`
	AppSettings AppSettings  `yaml:"app_settings"`
}

type AppSettings struct {
	URLLength    int    `yaml:"url_length" env:"URL_LENGTH" env-default:"10"`
	Storage      string `yaml:"storage" env:"STORAGE_TYPE" env-default:"map"`
	FirstURLPart string `yaml:"first_url_part" env:"FIRST_URL_PART" env-default:"https://somedomain.su/"`
}

type ServerConfig struct {
	Address string `yaml:"address" env:"SERVER_ADDRESS" env-default:":8080"`
}

type DBConfig struct {
	MigrationsDir string `yaml:"migrations_dir" env:"MIGRATIONS_DIR" env-default:"./migrations"`
	Conn          string `yaml:"conn" env:"POSTGRES_CONN" env-default:""`
}

func MustLoad(configPath string) Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatal("Cannot find config file")
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatal("Error while reading config")
	}

	return cfg
}
