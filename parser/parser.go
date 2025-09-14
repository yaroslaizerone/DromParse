package parser

import (
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"dromCrownParse/config"
	"dromCrownParse/models"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

var (
	idRe  = regexp.MustCompile(`(\d{5,})`)
	numRe = regexp.MustCompile(`(\d+)`)
)

// userAgents содержит список популярных User-Agent
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_0) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
}

// RandomUserAgent возвращает случайный User-Agent для имитации реального браузера
func RandomUserAgent() string {
	rand.Seed(time.Now().UnixNano())
	return userAgents[rand.Intn(len(userAgents))]
}

// FetchDocument загружает страницу по указанному URL и возвращает goquery.Document
// с учетом случайной задержки между запросами.
func FetchDocument(client *http.Client, urlStr string, cfg *config.Config) (*goquery.Document, error) {
	time.Sleep(time.Duration(cfg.DelayMin+rand.Intn(cfg.DelayMax-cfg.DelayMin+1)) * time.Second)

	req, _ := http.NewRequest("GET", urlStr, nil)
	req.Header.Set("User-Agent", RandomUserAgent())
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// ExtractPaginationInfo возвращает максимальное количество страниц и общее количество объявлений на странице.
func ExtractPaginationInfo(doc *goquery.Document) (maxPage int, totalCount int) {
	maxPage = 1
	totalCount = 0

	doc.Find(`a[data-ftid="component_pagination-item"]`).Each(func(i int, s *goquery.Selection) {
		txt := strings.TrimSpace(s.Text())
		if n, err := strconv.Atoi(txt); err == nil && n > maxPage {
			maxPage = n
		}
	})

	doc.Find(`a[data-ftid="bulls-list_bulls-tab"] span._3ynq47p`).Each(func(i int, s *goquery.Selection) {
		txt := strings.TrimSpace(s.Text())
		if m := numRe.FindStringSubmatch(txt); len(m) > 1 {
			if n, err := strconv.Atoi(m[1]); err == nil {
				totalCount = n
			}
		}
	})

	if totalCount == 0 {
		doc.Find(`div#tabs span`).Each(func(i int, s *goquery.Selection) {
			txt := strings.TrimSpace(s.Text())
			if m := numRe.FindStringSubmatch(txt); len(m) > 1 {
				if n, err := strconv.Atoi(m[1]); err == nil {
					totalCount = n
				}
			}
		})
	}

	return
}

// CollectAutosFromPage находит все объявления на странице и отправляет их в resultsCh.
// Проверяет уникальность ID и учитывает общее количество объявлений.
func CollectAutosFromPage(doc *goquery.Document, autos *[]models.Auto, mu *sync.Mutex, resultsCh chan<- models.Auto, collected *int, totalCount int) {
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		if *collected >= totalCount && totalCount > 0 {
			return
		}
		href, ok := s.Attr("href")
		if !ok {
			return
		}
		full := href
		base := doc.Url.Scheme + "://" + doc.Url.Host
		if strings.HasPrefix(href, "/") {
			full = base + href
		}
		m := idRe.FindStringSubmatch(full)
		if m == nil {
			return
		}
		id := m[1]

		mu.Lock()
		if !contains(*autos, id) && (*collected < totalCount || totalCount == 0) {
			*collected++
			mu.Unlock()
			auto := models.Auto{ID: id, URL: full}
			resultsCh <- auto
		} else {
			mu.Unlock()
		}
	})
}

// contains проверяет, есть ли авто с указанным ID в слайсе.
func contains(autos []models.Auto, id string) bool {
	for _, a := range autos {
		if a.ID == id {
			return true
		}
	}
	return false
}

// parseTextBySelector возвращает текст первого найденного элемента по селектору.
func parseTextBySelector(doc *goquery.Document, selector string) string {
	txt := strings.TrimSpace(doc.Find(selector).First().Text())
	if txt == "" {
		return "null"
	}
	return txt
}

// parseOwnText возвращает только собственный текст элемента (без детей) по селектору и индексу.
func parseOwnText(doc *goquery.Document, selector string, index int) string {
	txt := "null"
	doc.Find(selector).EachWithBreak(func(i int, s *goquery.Selection) bool {
		if i == index {
			txt = strings.TrimSpace(s.Clone().Children().Remove().End().Text())
			return false
		}
		return true
	})
	return txt
}

// ParseBrandModel извлекает бренд и модель автомобиля из хлебных крошек страницы.
func ParseBrandModel(doc *goquery.Document) (brand, model string) {
	doc.Find(`a[data-ftid="header_breadcrumb_link"] span._1lj8ai62`).Each(func(i int, s *goquery.Selection) {
		parent := s.Parent()
		if parent == nil {
			return
		}
		if val, exists := parent.Attr("data-ftid"); exists && val == "header_breadcrumb_link" {
			if payload, ok := parent.Attr("data-ga-stats-va-payload"); ok {
				if strings.Contains(payload, `"breadcrumb_number":3`) {
					brand = strings.TrimSpace(s.Text())
				} else if strings.Contains(payload, `"breadcrumb_number":4`) {
					model = strings.TrimSpace(s.Text())
				}
			}
		}
	})
	if brand == "" {
		brand = "null"
	}
	if model == "" {
		model = "null"
	}
	return
}
