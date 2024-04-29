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

func ScrapeDataFromLabirint(r repository.Repository, waitgroup *sync.WaitGroup) {

	c := colly.NewCollector(colly.Async(true))
	c.SetRequestTimeout(time.Minute * 20)

	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 20,
		RandomDelay: 240 * time.Millisecond,
	})

	pageToScrape := "https://www.labirint.ru/books/"
	vendor := "https://www.labirint.ru"
	regPages := regexp.MustCompile("Страниц: [0-9]+")
	regYear := regexp.MustCompile("[0-9]+")

	c.OnHTML("a.pagination-next__text", func(h *colly.HTMLElement) {
		c.Visit(h.Request.AbsoluteURL(h.Attr("href")))
	})

	c.OnHTML("div.inner-catalog", func(h *colly.HTMLElement) {

		h.DOM.Find("a.cover, genres-cover").Each(func(i int, s *goquery.Selection) {
			if strings.Contains(s.AttrOr("href", ""), "books") {
				c.Visit(h.Request.AbsoluteURL(s.AttrOr("href", "")))
			}
		})
	})

	c.OnHTML("a.b-menu-list-title", func(h *colly.HTMLElement) {
		c.Visit(h.Request.AbsoluteURL(h.Attr("href")))
	})

	c.OnHTML("div#product", func(h *colly.HTMLElement) {
		author := ""
		translator := ""
		currPrice, errCurr := strconv.Atoi(h.DOM.Find("div#product-info").AttrOr("data-discount-price", ""))
		oldPrice, _ := strconv.Atoi(h.DOM.Find("div#product-info").AttrOr("data-price", ""))

		title := strings.TrimSpace(h.DOM.Find("h1").Text())
		ageRestriction := h.DOM.Find("div#age_dopusk").Text()
		about := h.DOM.Find("div#product-about p").Text()

		imgPath := h.DOM.Find("img.book-img-cover,entered, loaded").AttrOr("data-src", "")
		productionSeries := h.DOM.Find("div.series a").Text()
		catgeory := h.DOM.Find("div.genre a").Text()
		publisher := h.DOM.Find("div.publisher a").Text()
		yearPublish := strings.Join(regYear.FindAllString(h.DOM.Find("div.publisher").Text(), -1), "")
		pageQuantity := strings.Join(regPages.FindAllString(h.DOM.Find("div.pages2").Text(), -1), "")

		pageQuantity = SafeSplit(pageQuantity, " ")
		format := SafeSplit(h.DOM.Find("div.dimensions").Text(), " ")
		weight := SafeSplit(h.DOM.Find("div.weight").Text(), " ")
		isbn := SafeSplit(h.DOM.Find("div.isbn").Text(), " ")

		h.DOM.Find("div.authors").Each(func(i int, s *goquery.Selection) {
			authors := strings.Split(s.Text(), ":")

			if len(authors) >= 2 {
				if authors[0] == "Автор" {
					author = authors[1]
				}
				if authors[0] == "Переводчик" {
					translator = authors[1]
				}
			}
		})

		if errCurr == nil && title != "" && currPrice != 0 {

			book := models.Book{
				CurrentPrice:     currPrice,
				OldPrice:         oldPrice,
				Title:            title,
				ImgPath:          imgPath,
				PageBookPath:     vendor + h.Request.URL.Path,
				VendorURL:        vendor,
				Vendor:           "Лабиринт",
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
				BookAbout:        about,
			}
			r.Db.Where("page_book_path = ?", book.PageBookPath).FirstOrCreate(&book)
		}

	})

	c.OnRequest(func(resp *colly.Request) {
		r.InfLog.Log("labirint", "Visiting: ", resp.URL.Path)
	})
	c.OnError(func(resp *colly.Response, err error) {
		r.ErrLog.Log("labirint", "Error while parsing web page", err.Error())
	})

	c.Visit(pageToScrape)
	c.Wait()
	waitgroup.Done()

}
