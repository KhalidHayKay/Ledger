package config

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Name        string
	Version     string
	Environment string
	Key         string
	Port        string
	Url         string
}

type RedisConfig struct {
	Host     string
	Port     string
	Addr     string
	Password string
}

type DBConfig struct {
	url string

	Connection string
	Host       string
	Port       string
	Database   string
	Username   string
	Password   string
}

func (db DBConfig) URL() string {
	if db.url != "" {
		return db.url
	}

	u := url.URL{
		Scheme: db.Connection,
		User:   url.UserPassword(db.Username, db.Password),
		Host:   fmt.Sprintf("%s:%s", db.Host, db.Port),
		Path:   db.Database,
	}

	return u.String()
}

type EnvType struct {
	App AppConfig

	DB DBConfig

	Redis RedisConfig

	BankAPIBaseURL string
}

var Env *EnvType

func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		log.Println(err)
	}

	Env = &EnvType{
		App: AppConfig{
			Name:        os.Getenv("APP_NAME"),
			Version:     os.Getenv("APP_VERSION"),
			Environment: os.Getenv("APP_ENV"),
			Key:         os.Getenv("APP_KEY"),
			Port:        os.Getenv("APP_PORT"),
			Url:         os.Getenv("APP_URL"),
		},

		DB: DBConfig{
			url: os.Getenv("DB_URL"),

			Connection: os.Getenv("DB_CONNECTION"),
			Host:       os.Getenv("DB_HOST"),
			Port:       os.Getenv("DB_PORT"),
			Database:   os.Getenv("DB_DATABASE"),
			Username:   os.Getenv("DB_USERNAME"),
			Password:   os.Getenv("DB_PASSWORD"),
		},

		Redis: RedisConfig{
			Host:     os.Getenv("REDIS_HOST"),
			Port:     os.Getenv("REDIS_PORT"),
			Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
			Password: os.Getenv("REDIS_PASSWORD"),
		},

		BankAPIBaseURL: os.Getenv("BANK_API_BASE_URL"),
	}

	return validate()
}
