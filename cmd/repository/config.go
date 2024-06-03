package repository

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HOST           string
	DB_PORT        string
	API_PORT       string
	PASSWORD       string
	USER           string
	DB             string
	SSLMODE        string
	EMAIL_USERNAME string
	EMAIL_PASSWORD string
}

func GetConfigVariables() (Config, error) {

	err := godotenv.Load("./db.env")

	if err != nil {
		return Config{}, err
	}

	config := Config{
		HOST:           os.Getenv("HOST"),
		DB_PORT:        os.Getenv("DB_PORT"),
		API_PORT:       os.Getenv("API_PORT"),
		PASSWORD:       os.Getenv("PASSWORD"),
		DB:             os.Getenv("DB_NAME"),
		USER:           os.Getenv("USER"),
		SSLMODE:        os.Getenv("SSLMODE_TYPE"),
		EMAIL_USERNAME: os.Getenv("EMAIL_USERNAME"),
		EMAIL_PASSWORD: os.Getenv("EMAIL_PASSWORD"),
	}

	return config, nil
}
