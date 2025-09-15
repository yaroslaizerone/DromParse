package fetch

import (
	"dromCrownParse/internal/models"
	"dromCrownParse/internal/parser"
	"dromCrownParse/internal/photos"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/EDDYCJY/fake-useragent"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
	"regexp"
	"strings"
	"sync"
)

func FetchDocument(url string) (*goquery.Document, error) {
	time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second)

	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", browser.Random())
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

func fetchAuto(url string, id string) models.Auto {
	doc, err := FetchDocument(url)
	if err != nil {
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

	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		th := strings.TrimSpace(s.Find("th").Text())
		td := strings.TrimSpace(s.Find("td").Text())
		parser.ParseTableRow(&a, th, td)
	})

	a.Photos = photos.DownloadPhotos(doc, id)

	return a
}

func ProcessPage(doc *goquery.Document, base string, idRe *regexp.Regexp, seen *map[string]bool, mu *sync.Mutex, collected *int, totalCount int, resultsCh chan models.Auto) {
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
			u, _ := url.Parse(base)
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
