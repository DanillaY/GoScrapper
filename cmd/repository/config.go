package repository

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	HOST            string
	DB_PORT         string
	API_PORT        string
	PASSWORD        string
	USER            string
	DB              string
	SSLMODE         string
	EMAIL_USERNAME  string
	EMAIL_PASSWORD  string
	EMAIL_SMTP      string
	EMAIL_SMTP_PORT int
}

func GetConfigVariables() (Config, error) {

	errEnv := godotenv.Load("./db.env")
	smtpPort, errPort := strconv.Atoi(os.Getenv("EMAIL_SMTP_PORT"))

	if errEnv != nil {
		return Config{}, errEnv
	}
	if errPort != nil {
		return Config{}, errPort
	}

	config := Config{
		HOST:            os.Getenv("HOST"),
		DB_PORT:         os.Getenv("DB_PORT"),
		API_PORT:        os.Getenv("API_PORT"),
		PASSWORD:        os.Getenv("PASSWORD"),
		DB:              os.Getenv("DB_NAME"),
		USER:            os.Getenv("USER"),
		SSLMODE:         os.Getenv("SSLMODE_TYPE"),
		EMAIL_USERNAME:  os.Getenv("EMAIL_USERNAME"),
		EMAIL_PASSWORD:  os.Getenv("EMAIL_PASSWORD"),
		EMAIL_SMTP:      os.Getenv("EMAIL_SMTP"),
		EMAIL_SMTP_PORT: smtpPort,
	}

	return config, nil
}
