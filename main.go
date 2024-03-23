package main

import (
	"log"
	"os"

	"github.com/DanillaY/GoScrapper/cmd/repository"
	"github.com/DanillaY/GoScrapper/cmd/task"
	slogger "github.com/jesse-rb/slogger-go"
	"github.com/joho/godotenv"
)

func main() {

	infoLogger := slogger.New(os.Stdout, slogger.ANSIMagenta, "info", log.Ldate)
	errorLogger := slogger.New(os.Stderr, slogger.ANSIRed, "error", log.Lshortfile+log.Ldate)

	err := godotenv.Load("./db.env")

	if err != nil {
		errorLogger.Log("main", "Error while getting env data", err.Error())
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
	repo := repository.Repository{Db: db, Config: &config, InfLog: infoLogger, ErrLog: errorLogger}

	if err != nil {
		errorLogger.Log("main", "Error while creating new connection with database", err.Error())
	}
	//waitgroup := &sync.WaitGroup{}
	//waitgroup.Add(1)

	task.ScrapeAllWebsites(repo)
}
