package task

import (
	"encoding/json"
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

func ScrapeDataFromRespulica(r repository.Repository, waitgroup *sync.WaitGroup) {

	c := colly.NewCollector(colly.Async(true))
	c.SetRequestTimeout(time.Minute * 20)
	c.Limit(&colly.LimitRule{
		Parallelism: 20,
		RandomDelay: 220 * time.Millisecond,
	})
	c.AllowURLRevisit = false

	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"
	pageToScrape := "https://www.respublica.ru/knigi?page=1"
	vendor := "https://www.respublica.ru"
	regNumbers := regexp.MustCompile("[0-9]+")
	priceTrim := strings.NewReplacer("\n", "", "\t", "", " ", "", " ", "")
	replacerAttr := strings.NewReplacer(
		"title", `"title"`,
		"values", `"values"`,
		"url", `"url"`,
		":a", `:"a"`,
		":t", `:"t"`,
		":u", `:"u"`,
		":s", `:"s"`,
		":r", `:"r"`,
		":P", `:"P"`,
		":q", `:"q"`,
		":O", `:"O"`,
		":Q", `:"Q"`,
		":R", `:"R"`,
		":S", `:"S"`,
		":T", `:"T"`,
		":U", `:"U"`,
		":v", `:"v"`,
		":V", `:"V"`,
		":p", `:"p"`,
		":N", `:"N"`,
		":n", `:"n"`,
		":M", `:"M"`,
		":o", `:"o"`,
		":w", `:"w"`,
		":m", `:"m"`,
		":j", `:"j"`,
		":k", `:"k"`,
		":W", `:"W"`,
		":l", `:"l"`,
		":h", `:"h"`,
		":i", `:"i"`,
		":X", `:"X"`,
		":L", `:"L"`,
		":K", `:"K"`,
		":Z", `:"Z"`,
	)

	c.OnHTML("a.pages-nav-link", func(h *colly.HTMLElement) {
		c.Visit(h.Request.AbsoluteURL(h.Attr("href")))
	})

	//Navigate to a description book page
	c.OnHTML("a.images, relative, icon-overlay", func(h *colly.HTMLElement) {
		c.Visit(h.Request.AbsoluteURL(h.Attr("href")))
	})

	//Fill book data and save to db
	c.OnHTML("html", func(h *colly.HTMLElement) {
		title := strings.TrimSpace(h.DOM.Find("h1.text-2xl, font-medium, text-gray-700, pb-3").Text())

		if title != "" {
			currPrice := priceTrim.Replace(h.DOM.Find("div.text-gray-700, font-medium, text-3xl").Text())
			oldPrice := priceTrim.Replace(h.DOM.Find("span.line-through").Text())
			about := h.DOM.Find("div.static-body").Text()
			imgPath := ""
			author := ""
			isbn := ""
			series := ""
			pagesQuantity := ""
			ageRestriction := ""
			yearPublished := ""
			bookCover := ""
			weight := ""

			scriptsJs := h.DOM.Find("script").Text()
			scriptsJs = SafeSplit(scriptsJs, "json_properties")[1:]

			jsonProductAttrs := strings.Split(scriptsJs, "}]}]")[0] + "}]}]"
			jsonProductAttrs = replacerAttr.Replace(jsonProductAttrs)

			var items []jsonObj
			err := json.Unmarshal([]byte(jsonProductAttrs), &items)

			if err != nil {
				r.ErrLog.Log("respublica", "Error while parsing json:", err)
			}

			for _, item := range items {

				if len(item.Values) >= 1 {
					switch item.Title {
					case "Серия":
						series = item.Values[0].Title
					case "Возрастные ограничения":
						ageRestriction = item.Values[0].Title
					case "Год издания":
						yearPublished = item.Values[0].Title
					case "Количество страниц":
						pagesQuantity = item.Values[0].Title
					case "Обложка":
						bookCover = item.Values[0].Title
					case "Вес, г":
						if item.Values[0].Title != "U" &&
							item.Values[0].Title != "o" &&
							item.Values[0].Title != "P" {
							weight = item.Values[0].Title
						}
					}

				}

			}

			h.DOM.Find("meta").Each(func(i int, s *goquery.Selection) {
				attr, exist := s.Attr("property")

				if exist {
					switch attr {
					case "og:image":
						imgPath = s.AttrOr("content", "")
					case "book:author":
						author = s.AttrOr("content", "")
					case "book:isbn":
						if s.AttrOr("content", "") != "undefined" {
							isbn = s.AttrOr("content", "")
						}
					}

				}
			})
			numberCurrPrice, errCurr := strconv.Atoi(regNumbers.FindString(currPrice))
			numberOldPrice, _ := strconv.Atoi(regNumbers.FindString(oldPrice))

			if currPrice != "" && imgPath != "" && errCurr == nil {

				book := models.Book{
					CurrentPrice:     numberCurrPrice,
					OldPrice:         numberOldPrice,
					Title:            strings.TrimSpace(title),
					ImgPath:          imgPath,
					PageBookPath:     vendor + h.Request.URL.Path,
					VendorURL:        vendor,
					Vendor:           "Республика",
					Author:           author,
					ProductionSeries: series,
					ISBN:             isbn,
					AgeRestriction:   ageRestriction,
					YearPublish:      yearPublished,
					PagesQuantity:    pagesQuantity,
					BookCover:        bookCover,
					Weight:           weight,
					BookAbout:        about,
				}
				r.Db.Where("page_book_path = ?", book.PageBookPath).FirstOrCreate(&book)
			}
		}
	})

	c.OnRequest(func(resp *colly.Request) {
		r.InfLog.Log("respublica", "Visiting: ", resp.URL.Path)
	})
	c.OnError(func(resp *colly.Response, err error) {
		r.ErrLog.Log("respublica", "Error while parsing web page", err.Error())
	})

	c.Visit(pageToScrape)
	c.Wait()
	waitgroup.Done()

}

type jsonObj struct {
	Title  string    `json:"title"`
	Values []subJson `json:"values"`
}
type subJson struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}
