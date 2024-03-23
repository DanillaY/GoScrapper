package task

import (
	"sync"

	"github.com/DanillaY/GoScrapper/cmd/models"
	"github.com/DanillaY/GoScrapper/cmd/repository"
)

func ScrapeAllWebsites(repo repository.Repository) {

	repo.Db.AutoMigrate(&models.Book{})

	waitgroup := &sync.WaitGroup{}
	waitgroup.Add(3)

	go ScrapeDataFromBook24(repo, waitgroup)
	go ScrapeDataFromVseSvobodny(repo, waitgroup)
	go ScrapeDataFromChitaiGorod(repo, waitgroup)
	waitgroup.Wait()
}

/*

	Bellow are common methods that are used in couple of parsers

*/

func CheckIfTheFieldExists(charBook map[string]string, key string) (value string) {
	val, exists := charBook[key]
	if !exists {
		val = ""
	}
	return val
}
