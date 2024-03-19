package main

import (
	"fmt"
	"os"

	"github.com/DanillaY/GoScrapper/cmd/repository"
	"github.com/DanillaY/GoScrapper/cmd/task"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load("./db.env")
	if err != nil {
		fmt.Println("Error while getting env data")
	}

	config := repository.Config{

		HOST:     os.Getenv("HOST"),
		DB_PORT:  os.Getenv("DB_PORT"),
		API_PORT: os.Getenv("API_PORT"),
		PASSWORD: os.Getenv("PASSWORD"),
		DB:       os.Getenv("DB_NAME"),
		USER:     os.Getenv("USER"),
		SSLMODE:  os.Getenv("SSLMODE_TYPE"),
	}

	db, err := repository.NewPostgresConnection(&config)
	repo := repository.Repository{Db: db, Config: &config}

	if err != nil {
		fmt.Println("Error while creating new connection with database")
	}

	task.ScrapeAllWebsites(config, repo.Db)
}
