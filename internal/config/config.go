package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerType string   `yaml:"server_type"`
	Port       string   `yaml:"port"`
	Storage    Storage  `yaml:"storage"`
	Logger     Logger   `yaml:"logger"`
	Security   Security `yaml:"security"`
}

type Storage struct {
	Type         string        `yaml:"type"`
	Connection   Connection    `yaml:"connection"`
	QueryTimeout time.Duration `yaml:"query_timeout"`
}
type Connection struct {
	Driver   string `yaml:"driver"`
	User     string
	Password string
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Database string `yaml:"database"`
}
type Logger struct {
	Type  string `yaml:"type"`
	Level string `yaml:"level"`
}

type Security struct {
	Hash     Hash     `yaml:"hash"`
	JWTToken JWTToken `yaml:"jwt_token"`
}

type Hash struct {
	Cost int `yaml:"cost"`
}

type JWTToken struct {
	SecretKey []byte
	Lifetime  time.Duration `yaml:"lifetime"`
	Prefix    string        `yaml:"prefix"`
}

var App *Config

func Init(path string) error {
	cfg, err := Load(path)
	if err != nil {
		return err
	}
	App = cfg
	return nil
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return &Config{}, err
	}
	var cfg Config
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return &Config{}, err
	}

	cfg.Storage.Connection.User = os.Getenv("DB_USER")
	cfg.Storage.Connection.Password = os.Getenv("DB_PASSWORD")
	cfg.Security.JWTToken.SecretKey = []byte(os.Getenv("JWT_SECRET"))

	return &cfg, nil
}
