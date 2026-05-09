package config

import (
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config представляет основную структуру конфигурации приложения.
type Config struct {
	ServerType string   `yaml:"server_type"`
	Port       string   `yaml:"port"`
	Storage    Storage  `yaml:"storage"`
	Logger     Logger   `yaml:"logger"`
	Tools      Tools    `yaml:"tools"`
	Security   Security `yaml:"security"`
}

// Storage представляет конфигурацию хранилища данных.
type Storage struct {
	Type          string        `yaml:"type"`
	Version       string        `yaml:"version"`
	ContainerName string        `yaml:"container_name"`
	Connection    Connection    `yaml:"connection"`
	QueryTimeout  time.Duration `yaml:"query_timeout"`
}

// Connection представляет параметры подключения к базе данных.
type Connection struct {
	Driver   string `yaml:"driver"`
	User     string
	Password string
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Database string `yaml:"database"`
}

// Logger представляет конфигурацию логирования.
type Logger struct {
	Type  string `yaml:"type"`
	Level string `yaml:"level"`
}

// Tools представляет конфигурацию вспомогательных инструментов.
type Tools struct {
	JSON string `yaml:"jsoncodec"`
}

// Security представляет конфигурацию безопасности.
type Security struct {
	Hash     Hash     `yaml:"hash"`
	JWTToken JWTToken `yaml:"jwt_token"`
}

// Hash представляет конфигурацию хэширования паролей.
type Hash struct {
	Cost int `yaml:"cost"`
}

// JWTToken представляет конфигурацию JWT токенов.
type JWTToken struct {
	SecretKey []byte
	Lifetime  time.Duration `yaml:"lifetime"`
	Prefix    string        `yaml:"prefix"`
}

// Load загружает и парсит конфигурацию из YAML файла по указанному пути.
func Load(path string) (*Config, error) {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		pathSlice := strings.Split(path, "/")
		path = pathSlice[len(pathSlice)-1]
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return &Config{}, err
	}
	var cfg Config
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return &Config{}, err
	}

	if _, err = os.Stat("/.dockerenv"); err == nil {
		cfg.Storage.Connection.Host = cfg.Storage.ContainerName
	}

	cfg.Storage.Connection.User = os.Getenv("DB_USER")
	cfg.Storage.Connection.Password = os.Getenv("DB_PASSWORD")
	cfg.Security.JWTToken.SecretKey = []byte(os.Getenv("JWT_SECRET"))

	return &cfg, nil
}
