package task

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DanillaY/GoScrapper/cmd/models"
	"github.com/DanillaY/GoScrapper/cmd/repository"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

func ScrapeDataFromBook24(r repository.Repository, waitgroup *sync.WaitGroup) {

	c := colly.NewCollector(colly.Async(true))
	c.SetRequestTimeout(time.Minute * 20)

	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"

	pageToScrape := "https://book24.ru/catalog"
	replacer := strings.NewReplacer("\n", "", "\t", "")
	regex := regexp.MustCompile("[0-9]+")
	vendor := "https://book24.ru"

	c.OnHTML("div.product-detail-page__body", func(c *colly.HTMLElement) {

		characteristicsBook := make(map[string]string)

		about := replacer.Replace(c.DOM.Find("div.product-about__text").Text())
		stockText := UnifyStockType(strings.TrimSpace(c.DOM.Find("button._block, b24-btn b24-btn__content").Text()))

		oldPrice := strings.TrimSpace(c.DOM.Find("span.product-sidebar-price__price-old").Text())
		currPrice := strings.TrimSpace(c.DOM.Find("span.product-sidebar-price__price").Text())

		if len(currPrice) > 0 {

			oldPriceArr := regex.FindAllString(strings.Replace(oldPrice, " ", "", -1), -1)
			currPriceArr := regex.FindAllString(strings.Replace(currPrice, " ", "", -1), -1)

			if len(oldPriceArr) > 0 {
				oldPrice = regex.FindAllString(strings.Replace(oldPrice, " ", "", -1), -1)[0]
			}
			if len(currPriceArr) > 0 {
				currPrice = regex.FindAllString(strings.Replace(currPrice, " ", "", -1), -1)[0]
			}
		}

		title := c.DOM.Find("h1.product-detail-page__title").Text()
		imgPath := c.DOM.Find("img.product-poster__main-image").AttrOr("src", "")

		c.DOM.Find("div.product-characteristic__item").Each(func(i int, s *goquery.Selection) {

			lines := strings.Split(strings.TrimSpace(s.Text()), "\n")

			if len(lines) > 0 {
				//key examples -> Автор, Серия, Переводчик
				key := strings.Split(lines[0], ":")[0]
				//value examples -> Ричард Бротиган, Романы-бротиганы, Гуревич Фаина
				value := strings.TrimSpace(strings.Split(lines[0], ":")[1])
				characteristicsBook[key] = value
			}

		})
		numberCurrPrice, errCurr := strconv.Atoi(currPrice)
		numberOldPrice, errOld := strconv.Atoi(oldPrice)

		if imgPath != "" &&
			((errCurr == nil && errOld == nil) || stockText != "Добавить в корзину") &&
			CheckIfTheFieldExists(characteristicsBook, "Автор") != "" {

			book := models.Book{
				CurrentPrice:     numberCurrPrice,
				OldPrice:         numberOldPrice,
				Title:            strings.TrimSpace(title),
				ImgPath:          imgPath,
				PageBookPath:     vendor + c.Request.URL.Path,
				VendorURL:        vendor,
				Vendor:           "Book24",
				Author:           CheckIfTheFieldExists(characteristicsBook, "Автор"),
				Translator:       CheckIfTheFieldExists(characteristicsBook, "Переводчик"),
				ProductionSeries: CheckIfTheFieldExists(characteristicsBook, "Серия"),
				Category:         strings.ReplaceAll(CheckIfTheFieldExists(characteristicsBook, "Раздел"), ",", " "),
				Publisher:        CheckIfTheFieldExists(characteristicsBook, "Издательство"),
				ISBN:             CheckIfTheFieldExists(characteristicsBook, "ISBN"),
				AgeRestriction:   CheckIfTheFieldExists(characteristicsBook, "Возрастное ограничение"),
				YearPublish:      CheckIfTheFieldExists(characteristicsBook, "Год издания"),
				PagesQuantity:    CheckIfTheFieldExists(characteristicsBook, "Количество страниц"),
				BookCover:        CheckIfTheFieldExists(characteristicsBook, "Переплет"),
				Format:           CheckIfTheFieldExists(characteristicsBook, "Формат"),
				Weight:           CheckIfTheFieldExists(characteristicsBook, "Вес"),
				InStockText:      stockText,
				BookAbout:        about,
			}

			if stockText == "Сообщить о поступлении" {
				book.CurrentPrice = 0
				book.OldPrice = 0
			}

			SaveBookAndNotifyUser(&r,
				numberCurrPrice, numberOldPrice,
				strings.TrimSpace(title), imgPath,
				vendor+c.Request.URL.Path,
				vendor, "Book24",
				CheckIfTheFieldExists(characteristicsBook, "Автор"), CheckIfTheFieldExists(characteristicsBook, "Переводчик"),
				CheckIfTheFieldExists(characteristicsBook, "Серия"), strings.ReplaceAll(CheckIfTheFieldExists(characteristicsBook, "Раздел"), ",", " "),
				CheckIfTheFieldExists(characteristicsBook, "Издательство"), CheckIfTheFieldExists(characteristicsBook, "ISBN"),
				CheckIfTheFieldExists(characteristicsBook, "Возрастные ограничения"),
				CheckIfTheFieldExists(characteristicsBook, "Год издания"), CheckIfTheFieldExists(characteristicsBook, "Количество страниц"),
				CheckIfTheFieldExists(characteristicsBook, "Размер"), CheckIfTheFieldExists(characteristicsBook, "Вес, г"),
				stockText, about)
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

	c.OnRequest(func(resp *colly.Request) {
		r.InfLog.Log("book24", "Visiting: ", resp.URL)
	})
	c.OnError(func(resp *colly.Response, err error) {
		r.ErrLog.Log("book24", "Error while parsing web page", err.Error())
	})

	c.Visit(pageToScrape)
	c.Wait()
	waitgroup.Done()
}
