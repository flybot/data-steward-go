package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type tokencfg struct {
	Secret          string `yaml:"secret" envconfig:"TOKEN_SECRET"`
	Lifetime        int    `yaml:"token_lifetime" envconfig:"TOKEN_LIFETIME"`
	RefreshLifetime int    `yaml:"refresh_token_lifetime" envconfig:"TOKEN_REFRESH_LIFETIME"`
}

type authentication struct {
	SignupAllowed   bool `yaml:"signup_allowed" envconfig:"SIGNUP_ALLOWED"`
	SignupApproving bool `yaml:"signup_approving" envconfig:"SIGNUP_APPROVING"`
}

type databasecfg struct {
	Host     string `yaml:"host" envconfig:"DB_HOST"`
	Database string `yaml:"database" envconfig:"DB_DATABASE"`
	User     string `yaml:"user" envconfig:"DB_USER"`
	Password string `yaml:"password" envconfig:"DB_PASS"`
	Port     int    `yaml:"port" envconfig:"DB_PORT"`
}

type redis struct {
	Host     string `yaml:"host" envconfig:"REDIS_HOST"`
	Password string `yaml:"password" envconfig:"REDIS_PASS"`
	Db       int    `yaml:"db" envconfig:"REDIS_DB"`
}

type webserver struct {
	Port   string `yaml:"port" envconfig:"WEBSRV_PORT"`
	UseTLS bool   `yaml:"tls" envconfig:"WEBSRV_TLS"`
}

type ServerConfig struct {
	WEB           webserver      `yaml:"webserver"`
	DB            databasecfg    `yaml:"database"`
	Redis         redis          `yaml:"redis"`
	Token         tokencfg       `yaml:"token"`
	Auth          authentication `yaml:"authentication"`
	SuperPassword string
}

func LoadConfig(fileName string) ServerConfig {
	var cfg ServerConfig
	readFile(&cfg, fileName)
	readEnv(&cfg)
	return cfg
}

func processError(err error) {
	fmt.Println(err)
	os.Exit(2)
}

func readFile(cfg *ServerConfig, fileName string) {
	f, err := os.Open(fileName)
	if err != nil {
		processError(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		processError(err)
	}
}

func readEnv(cfg *ServerConfig) {
	err := godotenv.Load()
	if err != nil {
		processError(err)
	}
	err = envconfig.Process("", cfg)
	if err != nil {
		processError(err)
	}
}
