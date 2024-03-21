package task

import (
	"github.com/DanillaY/GoScrapper/cmd/models"
	"github.com/DanillaY/GoScrapper/cmd/repository"
)

func ScrapeAllWebsites(repo repository.Repository) {

	repo.Db.AutoMigrate(&models.Book{})

	ScrapeDataFromBook24(repo)
	ScrapeDataFromVseSvobodny(repo)
}

func CheckIfTheFieldExists(charBook map[string]string, key string) string {
	val, exists := charBook[key]
	if !exists {
		val = ""
	}
	return val
}
