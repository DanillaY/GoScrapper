package task

import (
	db "github.com/DanillaY/GoScrapper/cmd/repository"
	"gorm.io/gorm"
)

func ScrapeAllWebsites(config db.Config, db *gorm.DB) {
	ScrapeDataFromBook24(config, db)
}
