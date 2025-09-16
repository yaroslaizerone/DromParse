package fetch

import (
	"dromCrownParse/internal/models"
	"dromCrownParse/internal/parser"
	"dromCrownParse/internal/photos"
	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

// FetchDocument делает HTTP GET-запрос по указанному URL и возвращает HTML-документ.
// В случае ошибок при запросе или парсинге документа возвращает ошибку.
func FetchDocument(url string) (*goquery.Document, error) {
	// Случайная задержка 1-3 секунды
	time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", browser.Random())

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Обработка кодировки страницы
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

// fetchAuto получает данные автомобиля с указанного URL и возвращает структуру models.Auto.
// Если FetchDocument возвращает ошибку, возвращается объект Auto с заполненными ID и URL.
func fetchAuto(url string, id string) models.Auto {
	doc, err := FetchDocument(url)
	if err != nil {
		log.Printf("Ошибка при загрузке документа %s: %v", url, err)
		return models.Auto{ID: id, URL: url}
	}

	a := models.Auto{ID: id, URL: url}

	brand, model := parser.ParseBrandModel(doc)
	a.Brand = brand
	a.Model = model
	a.Price = parser.ParsePrice(doc, "div.wb9m8q0")
	a.PriceMark = parser.ParseText(doc, `div[data-ga-stats-engine="ga|va"]`, 0)
	a.Generation = parser.ParseText(doc, `a[data-ga-stats-name="generation_link"]`, 0)
	a.Complectation = parser.ParseText(doc, `a[data-ga-stats-name="complectation_link"]`, 0)

	// Разбор таблиц с характеристиками
	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		th := strings.TrimSpace(s.Find("th").Text())
		td := strings.TrimSpace(s.Find("td").Text())
		parser.ParseTableRow(&a, th, td)
	})

	a.Photos = photos.DownloadPhotos(doc, id)

	return a
}

// ProcessPage обрабатывает HTML-документ страницы, извлекая ссылки на автомобили.
// Для каждой уникальной ссылки, подходящей под регулярное выражение idRe, вызывается fetchAuto.
// Найденные данные отправляются в resultsCh. Количество обработанных элементов ограничено totalCount.
// mu используется для безопасного обновления shared-срезов и счётчиков.
func ProcessPage(
	doc *goquery.Document,
	base string,
	idRe *regexp.Regexp,
	seen *map[string]bool,
	mu *sync.Mutex,
	collected *int,
	totalCount int,
	resultsCh chan models.Auto,
) {
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		if *collected >= totalCount && totalCount > 0 {
			return
		}

		href, ok := s.Attr("href")
		if !ok {
			return
		}

		full := href
		if strings.HasPrefix(href, "/") {
			u, err := url.Parse(base)
			if err != nil {
				log.Printf("Ошибка парсинга base URL %s: %v", base, err)
				return
			}
			full = u.Scheme + "://" + u.Host + href
		}

		m := idRe.FindStringSubmatch(full)
		if m == nil {
			return
		}
		id := m[1]

		mu.Lock()
		if !(*seen)[full] && *collected < totalCount {
			(*seen)[full] = true
			*collected++
			mu.Unlock()

			auto := fetchAuto(full, id)
			resultsCh <- auto
		} else {
			mu.Unlock()
		}
	})
}
