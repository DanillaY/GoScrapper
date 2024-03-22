package task

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DanillaY/GoScrapper/cmd/models"
	"github.com/DanillaY/GoScrapper/cmd/repository"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

func ScrapeDataFromChitaiGorod(r repository.Repository, waitgroup *sync.WaitGroup) {

	c := colly.NewCollector(colly.Async(true))
	c.SetRequestTimeout(time.Minute * 20)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 100,
		RandomDelay: 200 * time.Millisecond,
	})

	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"

	pageToScrape := "https://www.chitai-gorod.ru/catalog/books-18030?page=1"
	vendor := "https://www.chitai-gorod.ru"
	regex := regexp.MustCompile("[0-9]+")

	c.OnHTML("a.product-card__picture, product-card__row", func(h *colly.HTMLElement) {
		if !strings.Contains(h.Request.URL.Path, "product") {
			c.Visit(h.Request.AbsoluteURL(h.Attr("href")))
		}
	})

	c.OnHTML("meta", func(h *colly.HTMLElement) {
		if h.Attr("data-hid") == "og:url" && h.Attr("name") == "og:url" {
			url := h.Attr("content")
			urlParts := strings.Split(url, "=")

			if len(urlParts) >= 2 {
				pageNum, err := strconv.Atoi(urlParts[1])

				if err != nil {
					r.ErrLog.Log("chitaiGorod", "Error while parsing page number", err.Error())
				} else {
					fmt.Println(urlParts[0] + "=" + strconv.Itoa(pageNum+1))
					c.Visit(urlParts[0] + "=" + strconv.Itoa(pageNum+1))
				}
			}
		}
	})

	c.OnHTML("html", func(h *colly.HTMLElement) {

		oldPrice := formatPrice(regex, h, r, "span.product-offer-price__old-price")
		currPrice := formatPrice(regex, h, r, "span.product-offer-price__current")

		title := strings.TrimSpace(h.DOM.Find("h1.detail-product__header-title").Text())
		author := ""
		imgPath := ""

		h.DOM.Find("meta").Each(func(i int, s *goquery.Selection) {
			attr, exist := s.Attr("name")

			if attr == "og:image" && exist {
				imgPath = s.AttrOr("content", "")
			}
			if attr == "og:author" && exist {
				author = s.AttrOr("content", "")
			}
		})

		bookPath := h.Request.URL.String()

		characteristicsBook := make(map[string]string)

		h.DOM.Find("div.product-detail-features__item").Each(func(i int, s *goquery.Selection) {
			lines := strings.Split(strings.TrimSpace(s.Text()), "\n")
			keyValue := strings.Fields(strings.Join(lines, ""))

			doubleWordsKey := []string{"Количество", "Тип", "Год", "Возрастные", "Вес", "Вес,"}

			if len(keyValue) >= 2 && len(s.Text()) > 0 && len(lines) > 0 {

				if slices.Contains(doubleWordsKey, keyValue[0]) {
					characteristicsBook[keyValue[0]+" "+keyValue[1]] = keyValue[2]
				} else {
					characteristicsBook[keyValue[0]] = strings.Join(keyValue[1:], " ")
				}
			}

		})

		about := strings.TrimSpace(h.DOM.Find("article.detail-description__text").Text())
		about = strings.Replace(about, "\t", "", -1)

		if about != "" && currPrice != 0 && title != "" && bookPath != "" {
			book := models.Book{
				CurrentPrice:     currPrice,
				OldPrice:         oldPrice,
				Title:            strings.TrimSpace(title),
				ImgPath:          imgPath,
				PageBookPath:     bookPath,
				Vendor:           vendor,
				Category:         "",
				Author:           author,
				Translator:       CheckIfTheFieldExists(characteristicsBook, "Переводчик"),
				ProductionSeries: CheckIfTheFieldExists(characteristicsBook, "Серия"),
				Publisher:        CheckIfTheFieldExists(characteristicsBook, "Издательство"),
				ISBN:             CheckIfTheFieldExists(characteristicsBook, "ISBN"),
				AgeRestriction:   CheckIfTheFieldExists(characteristicsBook, "Возрастные ограничения"),
				YearPublish:      CheckIfTheFieldExists(characteristicsBook, "Год издания"),
				PagesQuantity:    CheckIfTheFieldExists(characteristicsBook, "Количество страниц"),
				BookCover:        CheckIfTheFieldExists(characteristicsBook, "Тип обложки"),
				Format:           CheckIfTheFieldExists(characteristicsBook, "Размер"),
				Weight:           CheckIfTheFieldExists(characteristicsBook, "Вес, г"),
				BookAbout:        strings.TrimSpace(about),
			}
			r.Db.Create(&book)
		}
	})

	c.OnRequest(func(resp *colly.Request) {
		resp.Headers.Set("User-Agent", "1 Mozilla/5.0 (iPad; CPU OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148")
		r.InfLog.Log("chitaiGorod", "Visiting: ", resp.URL.Path)
	})
	c.OnError(func(resp *colly.Response, err error) {
		r.ErrLog.Log("chitaiGorod", "Error while parsing web page", err.Error())
	})

	c.Visit(pageToScrape)
	c.Wait()
	waitgroup.Done()

}

func formatPrice(regex *regexp.Regexp, h *colly.HTMLElement, r repository.Repository, element string) (price int) {
	priceStr := strings.Replace(h.DOM.Find(element).Text(), " ", "", -1)
	tmpslice := regex.FindAllString(priceStr, -1)

	if len(tmpslice) > 0 {
		price, err := strconv.Atoi(tmpslice[0])
		if err != nil {
			r.ErrLog.Log("chitaiGorod", "Could not parse price", err.Error())
		}
		return price
	}
	return 0
}
