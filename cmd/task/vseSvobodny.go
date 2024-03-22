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

func ScrapeDataFromVseSvobodny(r repository.Repository, waitgroup *sync.WaitGroup) {

	c := colly.NewCollector(colly.Async(false))
	c.SetRequestTimeout(time.Minute * 20)
	c.AllowURLRevisit = false
	pagetoScrape := "https://vse-svobodny.com/shop/"

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 100,
		RandomDelay: 300 * time.Millisecond,
	})

	vendor := "https://vse-svobodny.com/"
	regex := regexp.MustCompile("[0-9]+")

	c.OnHTML("li.product", func(h *colly.HTMLElement) {
		if !strings.Contains(h.Request.URL.Path, "product") {
			c.Visit(h.Request.AbsoluteURL((h.ChildAttr("a", "href"))))
		}
	})

	c.OnHTML("a.next", func(h *colly.HTMLElement) {
		c.Visit(h.Attr("href"))
	})

	c.OnHTML("div.ast-woocommerce-container, product-type-simple, product, type-product, ast-article-single",
		func(h *colly.HTMLElement) {

			characteristicsBook := make(map[string]string)
			about := h.DOM.Find("div.woocommerce-Tabs-panel--description").Text()

			currPrice := strings.Join(regex.FindAllString(h.DOM.Find("p.price").Text(), -1), "")

			currPriceNumber := 0
			if len(currPrice) > 0 {
				price, err := strconv.Atoi(currPrice)
				if err != nil {
					r.ErrLog.Log("vseSvobodny", "Error while parsing price value", err.Error()+" page: "+h.Request.URL.Path)
				} else {
					currPriceNumber = price
				}
			}

			title := h.DOM.Find("h1.product_title, entry-title").Text()
			imgPath := h.DOM.Find("img.wp-post-image").AttrOr("src", "")
			bookPath := h.DOM.Find("form.cart").AttrOr("action", "")
			category := h.DOM.Find("span.posted_in").Text()

			if strings.Contains(category, ":") {
				category = strings.TrimSpace(strings.Split(category, ":")[1])
			} else {
				category = ""
			}

			h.DOM.Find("tr.woocommerce-product-attributes-item").Each(func(i int, s *goquery.Selection) {
				lines := strings.Split(strings.TrimSpace(s.Text()), "\n")

				keyValue := strings.Fields(strings.Join(lines, ""))
				if len(keyValue) >= 2 && len(s.Text()) > 0 {
					if keyValue[0] == "Количество" {
						characteristicsBook[keyValue[0]+" "+keyValue[1]] = keyValue[2]
					} else {
						characteristicsBook[keyValue[0]] = strings.Join(keyValue[1:], " ")
					}
				}
			})

			if about != "" && currPriceNumber != 0 && title != "" {
				book := models.Book{
					CurrentPrice:  currPriceNumber,
					OldPrice:      0,
					Title:         strings.TrimSpace(title),
					ImgPath:       imgPath,
					PageBookPath:  bookPath,
					Vendor:        vendor,
					Category:      category,
					Author:        CheckIfTheFieldExists(characteristicsBook, "Автор"),
					Translator:    CheckIfTheFieldExists(characteristicsBook, "Переводчик"),
					Publisher:     CheckIfTheFieldExists(characteristicsBook, "Издательство"),
					YearPublish:   CheckIfTheFieldExists(characteristicsBook, "Год"),
					PagesQuantity: CheckIfTheFieldExists(characteristicsBook, "Количество страниц"),
					BookCover:     CheckIfTheFieldExists(characteristicsBook, "Переплёт"),
					BookAbout:     strings.TrimSpace(about),
				}
				r.Db.Create(&book)
			}
		})

	c.OnRequest(func(resp *colly.Request) {
		resp.Headers.Set("User-Agent", "1 Mozilla/5.0 (iPad; CPU OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148")
		r.InfLog.Log("vseSvobodny", "Visiting: ", resp.URL.Path)
	})
	c.OnError(func(resp *colly.Response, err error) {
		r.ErrLog.Log("vseSvobodny", "Error while parsing web page", err.Error())
	})

	c.Visit(pagetoScrape)
	c.Wait()
	waitgroup.Done()
}
