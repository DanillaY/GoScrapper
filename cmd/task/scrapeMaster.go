package task

import (
	"strings"
	"sync"

	"github.com/DanillaY/GoScrapper/cmd/models"
	"github.com/DanillaY/GoScrapper/cmd/repository"
)

func ScrapeAllWebsites(repo repository.Repository) {

	repo.Db.AutoMigrate(&models.Book{})

	waitgroup := &sync.WaitGroup{}
	waitgroup.Add(5)

	go ScrapeDataFromBook24(repo, waitgroup)
	go ScrapeDataFromVseSvobodny(repo, waitgroup)
	go ScrapeDataFromChitaiGorod(repo, waitgroup)
	go ScrapeDataFromRespulica(repo, waitgroup)
	go ScrapeDataFromLabirint(repo, waitgroup)
	waitgroup.Wait()
}

// common method that is used in a couple of parsers
func CheckIfTheFieldExists(charBook map[string]string, key string) (value string) {
	val, exists := charBook[key]
	if !exists {
		val = ""
	}
	return val
}

func SafeSplit(str string, separator string) string {
	if strings.Contains(str, separator) {
		str = strings.Split(str, separator)[1]
	}
	return str
}
