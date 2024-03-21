package repository

import (
	"fmt"

	slogger "github.com/jesse-rb/slogger-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	Db     *gorm.DB
	Config *Config
	InfLog *slogger.Logger
	ErrLog *slogger.Logger
}

func NewPostgresConnection(c *Config) (db *gorm.DB, e error) {
	db, err := gorm.Open(postgres.Open(
		"host="+c.HOST+
			" port="+c.DB_PORT+
			" password="+c.PASSWORD+
			" user="+c.USER+
			" dbname="+c.DB+
			" sslmode="+c.SSLMODE), &gorm.Config{})
	if err != nil {
		fmt.Println("Error while opening the connection to database")
		return db, err
	}
	return db, nil
}
