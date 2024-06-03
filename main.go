package main

import (
	"log"
	"os"

	"github.com/DanillaY/GoScrapper/cmd/repository"
	"github.com/DanillaY/GoScrapper/cmd/task"
	slogger "github.com/jesse-rb/slogger-go"
)

func main() {

	infoLogger := slogger.New(os.Stdout, slogger.ANSIMagenta, "info", log.Ldate)
	errorLogger := slogger.New(os.Stderr, slogger.ANSIRed, "error", log.Lshortfile+log.Ldate)

	config, err := repository.GetConfigVariables()

	if err != nil {
		errorLogger.Log("main", "Could not load env file", err)
	}

	db, err := repository.NewPostgresConnection(&config)
	repo := repository.Repository{Db: db, Config: &config, InfLog: infoLogger, ErrLog: errorLogger}

	if err != nil {
		errorLogger.Log("main", "Error while creating new connection with database", err.Error())
	}

	task.ScrapeAllWebsites(repo)
}
