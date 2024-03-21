package task

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/DanillaY/GoScrapper/cmd/models"
	"github.com/DanillaY/GoScrapper/cmd/repository"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

func ScrapeDataFromVseSvobodny(r repository.Repository) {

	c := colly.NewCollector(colly.Async(false))
	c.SetRequestTimeout(time.Minute * 20)
	c.AllowURLRevisit = false
	pagetoScrape := "https://vse-svobodny.com/shop/"

	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 3,
		RandomDelay: 350 * time.Millisecond,
	})
	c.WithTransport(&http.Transport{
		DisableKeepAlives: true,
	})

	vendor := "https://vse-svobodny.com/"
	regex := regexp.MustCompile("[0-9]+")

	c.OnHTML("li.product", func(h *colly.HTMLElement) {
		c.Visit(h.Request.AbsoluteURL((h.ChildAttr("a", "href"))))
	})

	c.OnHTML("a.next", func(h *colly.HTMLElement) {
		//TO DO: figure out how to change navigation with pages
		//c.Visit(h.Attr("href"))
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

			h.DOM.Find("tr.woocommerce-product-attributes-item").Each(func(i int, s *goquery.Selection) {
				if len(s.Text()) > 0 {
					lines := strings.Split(strings.TrimSpace(s.Text()), "\n")

					if len(lines) > 0 {
						keyValue := strings.Fields(strings.Join(lines, ""))
						if len(keyValue) >= 2 {
							if keyValue[0] == "Количество" {
								characteristicsBook[keyValue[0]+" "+keyValue[1]] = keyValue[2]
							} else {
								characteristicsBook[keyValue[0]] = strings.Join(keyValue[1:], " ")
							}
						}

					}
				}

			})

			if about != "" && currPriceNumber != 0 && title != "" {
				book := models.Book{
					CurrentPrice:     currPriceNumber,
					OldPrice:         0,
					Title:            strings.TrimSpace(title),
					ImgPath:          imgPath,
					PageBookPath:     bookPath,
					Vendor:           vendor,
					Author:           CheckIfTheFieldExists(characteristicsBook, "Автор"),
					Translator:       CheckIfTheFieldExists(characteristicsBook, "Переводчик"),
					ProductionSeries: CheckIfTheFieldExists(characteristicsBook, "Серия"),
					Category:         CheckIfTheFieldExists(characteristicsBook, "Раздел"),
					Publisher:        CheckIfTheFieldExists(characteristicsBook, "Издательство"),
					YearPublish:      CheckIfTheFieldExists(characteristicsBook, "Год"),
					PagesQuantity:    CheckIfTheFieldExists(characteristicsBook, "Количество страниц"),
					BookCover:        CheckIfTheFieldExists(characteristicsBook, "Переплёт"),
					BookAbout:        strings.TrimSpace(about),
				}
				fmt.Println("---------" + book.Title)
				r.Db.Create(&book)
			}

			c.Visit(h.Attr("href"))
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
}