package task

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/DanillaY/GoScrapper/cmd/models"
	db "github.com/DanillaY/GoScrapper/cmd/repository"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"gorm.io/gorm"
)

func ScrapeDataFromBook24(config db.Config, db *gorm.DB) {

	db.AutoMigrate(&models.Book{})

	c := colly.NewCollector(colly.Async(true))
	c.SetRequestTimeout(time.Minute * 20)
	//c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 2})
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"

	pageToScrape := "https://book24.ru/catalog/"
	replacer := strings.NewReplacer("\n", "", "\t", "")
	re := regexp.MustCompile("[0-9]+")
	vendor := "https://book24.ru"

	c.OnHTML("div.product-detail-page__body", func(c *colly.HTMLElement) {

		characteristicsBook := make(map[string]string)

		about := replacer.Replace(c.DOM.Find("div.product-about__text").Text())

		oldPrice := strings.TrimSpace(c.DOM.Find("span.product-sidebar-price__price-old").Text())
		currPrice := strings.TrimSpace(c.DOM.Find("span.product-sidebar-price__price").Text())

		if len(currPrice) > 0 {

			oldPriceArr := re.FindAllString(strings.Replace(oldPrice, " ", "", -1), -1)
			currPriceArr := re.FindAllString(strings.Replace(currPrice, " ", "", -1), -1)

			if len(oldPriceArr) > 0 {
				oldPrice = re.FindAllString(strings.Replace(oldPrice, " ", "", -1), -1)[0]
			}
			if len(currPriceArr) > 0 {
				currPrice = re.FindAllString(strings.Replace(currPrice, " ", "", -1), -1)[0]
			}
		}

		title := c.DOM.Find("h1.product-detail-page__title").Text()
		imgPath := c.DOM.Find("img.product-poster__main-image").AttrOr("src", "")

		c.DOM.Find("div.product-characteristic__item").Each(func(i int, s *goquery.Selection) {

			if len(s.Text()) > 0 {
				lines := strings.Split(strings.TrimSpace(s.Text()), "\n")

				if len(lines) > 0 {
					//key examples -> Автор, Серия, Переводчик
					key := strings.Split(lines[0], ":")[0]
					//value examples -> Ричард Бротиган, Романы-бротиганы, Гуревич Фаина
					value := strings.TrimSpace(strings.Split(lines[0], ":")[1])
					characteristicsBook[key] = value
				}
			}
		})
		numberCurrPrice, errCurr := strconv.Atoi(currPrice)
		numberOldPrice, errOld := strconv.Atoi(oldPrice)

		if imgPath != "" && errCurr == nil && errOld == nil {

			book := models.Book{
				CurrentPrice:     numberCurrPrice,
				OldPrice:         numberOldPrice,
				Title:            strings.TrimSpace(title),
				ImgPath:          imgPath,
				PageBookPath:     vendor + c.Request.URL.Path,
				Vendor:           vendor,
				Author:           checkIfTheFieldExists(characteristicsBook, "Автор"),
				Translator:       checkIfTheFieldExists(characteristicsBook, "Переводчик"),
				ProductionSeries: checkIfTheFieldExists(characteristicsBook, "Серия"),
				Category:         checkIfTheFieldExists(characteristicsBook, "Раздел"),
				Publisher:        checkIfTheFieldExists(characteristicsBook, "Издательство"),
				ISBN:             checkIfTheFieldExists(characteristicsBook, "ISBN"),
				AgeRestriction:   checkIfTheFieldExists(characteristicsBook, "Возрастное ограничение"),
				YearPublish:      checkIfTheFieldExists(characteristicsBook, "Год издания"),
				PagesQuantity:    checkIfTheFieldExists(characteristicsBook, "Количество страниц"),
				BookCover:        checkIfTheFieldExists(characteristicsBook, "Переплет"),
				Format:           checkIfTheFieldExists(characteristicsBook, "Формат"),
				Weight:           checkIfTheFieldExists(characteristicsBook, "Вес"),
				BookAbout:        about,
			}
			db.Create(&book)
		}
	})
	c.OnHTML("div.product-card__image-holder", func(e *colly.HTMLElement) {
		c.Visit(e.Request.AbsoluteURL((e.ChildAttr("a", "href"))))
	})

	c.OnHTML("link", func(e *colly.HTMLElement) {
		if e.Attr("rel") == "next" {
			c.Visit(e.Request.AbsoluteURL(e.Attr("href")))
		}
	})

	c.OnScraped(func(response *colly.Response) {

	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Could not get param: " + err.Error())
	})

	c.Visit(pageToScrape)
	c.Wait()
}

func checkIfTheFieldExists(charBook map[string]string, key string) string {
	val, exists := charBook[key]
	if !exists {
		val = ""
	}
	return val
}
