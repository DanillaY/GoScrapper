package task

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DanillaY/GoScrapper/cmd/repository"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

func ScrapeDataFromChitaiGorod(r repository.Repository, waitgroup *sync.WaitGroup) {

	c := colly.NewCollector(colly.Async(true))
	c.SetRequestTimeout(time.Minute * 20)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 13,
		RandomDelay: 250 * time.Millisecond,
	})

	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"

	pageToScrape := "https://www.chitai-gorod.ru/catalog/books-18030?page=1"
	vendor := "https://www.chitai-gorod.ru"
	regex := regexp.MustCompile("[0-9]+")
	replacer := strings.NewReplacer("\n", "", "\t", "", " ", "")

	c.OnHTML("a.product-card__picture, product-card__row", func(h *colly.HTMLElement) {
		if !strings.Contains(h.Request.URL.Path, "product") {
			c.Visit(h.Request.AbsoluteURL(h.Attr("href")))
		}
	})

	c.OnHTML("a.pagination__button", func(h *colly.HTMLElement) {
		if strings.Contains(h.Request.URL.Path, "catalog") {
			c.Visit(h.Request.AbsoluteURL(h.Attr("href")))
		}
	})

	c.OnHTML("html", func(h *colly.HTMLElement) {

		if strings.Contains(h.Request.URL.String(), "product") {
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

			category := strings.Fields(h.DOM.Find("ul.product-breadcrumbs, detail-product__breadcrumbs").Text())
			category = slices.Compact(category)

			bookPath := h.Request.URL.String()
			stockText := UnifyStockType(SafeSplit(replacer.Replace(h.DOM.Find("button.product-offer-button").Find("div.chg-app-button__content").Text()), "  ", 1))

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
			yearPublish, errYear := strconv.Atoi(CheckIfTheFieldExists(characteristicsBook, "Год издания"))
			if errYear != nil {
				yearPublish = 0
			}

			if about != "" && currPrice != 0 && title != "" && bookPath != "" && len(category) > 1 {
				SaveBookAndNotifyUser(&r,
					currPrice, oldPrice,
					title, imgPath,
					vendor+h.Request.URL.Path,
					vendor, "Читай город",
					author, CheckIfTheFieldExists(characteristicsBook, "Переводчик"),
					CheckIfTheFieldExists(characteristicsBook, "Серия"), strings.ReplaceAll(strings.Join(category[1:], " "), ",", " "),
					CheckIfTheFieldExists(characteristicsBook, "Издательство"), CheckIfTheFieldExists(characteristicsBook, "ISBN"),
					CheckIfTheFieldExists(characteristicsBook, "Возрастные ограничения"),
					yearPublish, "",
					CheckIfTheFieldExists(characteristicsBook, "Размер"), CheckIfTheFieldExists(characteristicsBook, "Вес, г"),
					stockText, about)
			}
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
