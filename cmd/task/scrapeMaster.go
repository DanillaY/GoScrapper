package task

import (
	"strings"
	"sync"

	"github.com/DanillaY/GoScrapper/cmd/models"
	"github.com/DanillaY/GoScrapper/cmd/repository"
)

func ScrapeAllWebsites(repo repository.Repository) {

	repo.Db.AutoMigrate(&models.Book{})
	repo.Db.AutoMigrate(&models.User{})
	repo.Db.AutoMigrate(&models.BookUser{})
	errUser := repo.Db.SetupJoinTable(&models.Book{}, "User", &models.BookUser{})
	errBook := repo.Db.SetupJoinTable(&models.User{}, "Book", &models.BookUser{})

	if errUser == nil && errBook == nil {

		waitgroup := &sync.WaitGroup{}
		waitgroup.Add(5)

		go ScrapeDataFromBook24(repo, waitgroup)
		go ScrapeDataFromVseSvobodny(repo, waitgroup)
		go ScrapeDataFromChitaiGorod(repo, waitgroup)
		go ScrapeDataFromRespulica(repo, waitgroup)
		go ScrapeDataFromLabirint(repo, waitgroup)
		waitgroup.Wait()

	} else {
		repo.ErrLog.Log("Scrape master", "Could not bind user and books in sql table", errBook)
	}

}

// common methods that is used in a couple of parsers
func CheckIfTheFieldExists(charBook map[string]string, key string) (value string) {
	val, exists := charBook[key]
	if !exists {
		val = ""
	}
	return val
}

func SafeSplit(str string, separator string, indexGet int) string {
	if strings.Contains(str, separator) {
		str = strings.Split(str, separator)[indexGet]
	}
	return str
}

func UnifyStockType(stockText string) string {
	var result string

	stockText = strings.ToLower(stockText)
	if stockText == "ожидается" || strings.Contains(stockText, "предзаказ") {
		result = "Ожидается"
	} else if stockText == "добавить в корзину" || stockText == "купить" || stockText == "на складе" {
		result = "В наличии"
	} else if stockText != "false" {
		result = "Нет наличии"
	}

	return result
}

func SaveBookAndNotifyUser(
	r *repository.Repository,
	currPrice int,
	oldPrice int,
	title string,
	imgPath string,
	pageBookPath string,
	vendorUrl string,
	vendor string,
	author string,
	translator string,
	productionSeries string,
	catgeory string,
	publisher string,
	isbn string,
	ageRestriction string,
	yearPublish string,
	pageQuantity string,
	format string,
	weight string,
	stockText string,
	about string,
) {
	book := models.Book{
		CurrentPrice:     currPrice,
		OldPrice:         oldPrice,
		Title:            title,
		ImgPath:          imgPath,
		PageBookPath:     pageBookPath,
		VendorURL:        vendorUrl,
		Vendor:           vendor,
		Author:           strings.TrimSpace(author),
		Translator:       translator,
		ProductionSeries: productionSeries,
		Category:         catgeory,
		Publisher:        publisher,
		ISBN:             isbn,
		AgeRestriction:   ageRestriction,
		YearPublish:      yearPublish,
		PagesQuantity:    pageQuantity,
		Format:           format,
		Weight:           weight,
		InStockText:      stockText,
		BookAbout:        about,
	}

	bookOld := models.Book{}
	err := r.Db.Preload("User").Find(&bookOld, "page_book_path = ?", pageBookPath).Error

	if bookOld.PageBookPath != "" {
		r.Db.Model(&bookOld).Where("page_book_path = ?", bookOld.PageBookPath).
			Updates(map[string]interface{}{
				"current_price":     currPrice,
				"old_price":         oldPrice,
				"title":             title,
				"img_path":          imgPath,
				"page_book_path":    pageBookPath,
				"vendor_url":        vendorUrl,
				"vendor":            vendor,
				"author":            strings.TrimSpace(author),
				"translator":        translator,
				"production_series": productionSeries,
				"category":          catgeory,
				"publisher":         publisher,
				"isbn":              isbn,
				"age_restriction":   ageRestriction,
				"year_publish":      yearPublish,
				"pages_quantity":    pageQuantity,
				"book_cover":        format,
				"weight":            weight,
				"is_in_stock":       stockText,
				"book_about":        about,
			})
	} else if err == nil {
		r.Db.Create(&book)
	}
}
