package dromCrownParse

import (
	"encoding/csv"
	"fmt"
	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Auto struct {
	ID            string
	URL           string
	Brand         string
	Model         string
	Price         string
	PriceMark     string
	Generation    string
	Complectation string
	Mileage       string
	NoMileageRF   string
	Color         string
	BodyType      string
	Power         string
	FuelType      string
	EngineVolume  string
	Photos        []string
}

func main() {
	cities := []string{
		"https://vladivostok.drom.ru/auto/all/",
		"https://ussuriisk.drom.ru/auto/all/",
	}

	params := url.Values{}
	params.Add("multiselect[]", "9_4_15_all")
	params.Add("multiselect[]", "9_4_16_all")
	params.Add("ph", "1")
	params.Add("pts", "2")
	params.Add("damaged", "2")
	params.Add("unsold", "1")
	params.Add("whereabouts[]", "0")

	idRe := regexp.MustCompile(`(\d{5,})`)
	numRe := regexp.MustCompile(`(\d+)`)

	seen := make(map[string]bool)
	var mu sync.Mutex
	var autos []Auto
	maxWorkers := 5

	for _, base := range cities {
		fmt.Printf("\n=== Сканирую город: %s ===\n", base)
		firstURL := base + "?" + params.Encode()

		doc, err := fetchDocument(firstURL)
		if err != nil {
			log.Printf("Ошибка первого запроса: %v", err)
			continue
		}

		maxPage := 1
		doc.Find(`a[data-ftid="component_pagination-item"]`).Each(func(i int, s *goquery.Selection) {
			txt := strings.TrimSpace(s.Text())
			if n, err := strconv.Atoi(txt); err == nil && n > maxPage {
				maxPage = n
			}
		})

		totalCount := 0
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

		fmt.Printf("Всего страниц: %d, объявлений: %d\n", maxPage, totalCount)

		collected := 0
		pagesCh := make(chan int, maxPage)
		resultsCh := make(chan Auto, totalCount)
		var wg sync.WaitGroup

		for page := 1; page <= maxPage; page++ {
			pagesCh <- page
		}
		close(pagesCh)

		for w := 0; w < maxWorkers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for page := range pagesCh {
					if collected >= totalCount && totalCount > 0 {
						return
					}
					pageURL := fmt.Sprintf("%spage%d/?%s", base, page, params.Encode())
					doc, err := fetchDocument(pageURL)
					if err != nil {
						continue
					}

					doc.Find("a").Each(func(i int, s *goquery.Selection) {
						if collected >= totalCount && totalCount > 0 {
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
						if !seen[full] && collected < totalCount {
							seen[full] = true
							collected++
							mu.Unlock()
							auto := fetchAuto(full, id)
							resultsCh <- auto
						} else {
							mu.Unlock()
						}
					})
				}
			}()
		}

		wg.Wait()
		close(resultsCh)

		for a := range resultsCh {
			autos = append(autos, a)
		}
	}

	saveResults(autos)
	fmt.Printf("\nГотово: %d автомобилей\n", len(autos))
}

// fetchDocument получает HTML и конвертирует в UTF-8
func fetchDocument(url string) (*goquery.Document, error) {
	// случайная задержка 1–3 секунды
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

// fetchAuto парсит данные одного объявления
func fetchAuto(url string, id string) Auto {
	doc, err := fetchDocument(url)
	if err != nil {
		return Auto{ID: id, URL: url}
	}

	a := Auto{ID: id, URL: url}

	// --- Парсим основные поля ---
	brand, model := parseBrandModel(doc)
	a.Brand = brand
	a.Model = model
	a.Price = parsePrice(doc, "div.wb9m8q0")
	a.PriceMark = parseText(doc, `div[data-ga-stats-engine="ga|va"]`, 0)
	a.Generation = parseText(doc, `a[data-ga-stats-name="generation_link"]`, 0)
	a.Complectation = parseText(doc, `a[data-ga-stats-name="complectation_link"]`, 0)

	// --- Парсим таблицу свойств ---
	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		th := strings.TrimSpace(s.Find("th").Text())
		td := strings.TrimSpace(s.Find("td").Text())

		switch th {
		case "Пробег":
			parts := strings.Split(td, ",")
			a.Mileage = strings.TrimSpace(strings.ReplaceAll(parts[0], "км", ""))
			if strings.Contains(td, "без") {
				a.NoMileageRF = "да"
			} else {
				a.NoMileageRF = "нет"
			}

		case "Цвет":
			a.Color = cleanText(td)
		case "Тип кузова":
			a.BodyType = td
		case "Двигатель":
			parts := strings.Split(td, ",")
			if len(parts) >= 2 {
				a.FuelType = strings.TrimSpace(parts[0])
				a.EngineVolume = strings.TrimSpace(parts[1])
			}
		case "Мощность":
			// создаём строку с реальным неразрывным пробелом \u00A0
			nbsp := "\u00A0"
			pattern := `(\d+)[\s` + nbsp + `]*л\.с`

			re := regexp.MustCompile(pattern)
			match := re.FindStringSubmatch(td)
			if len(match) > 1 {
				a.Power = match[1]
			} else {
				a.Power = "null"
			}
		}
	})

	// --- Скачиваем фотографии ---
	a.Photos = downloadPhotos(doc, id)

	return a
}

func parseOwnText(doc *goquery.Document, selector string, index int) string {
	txt := "null"
	doc.Find(selector).EachWithBreak(func(i int, s *goquery.Selection) bool {
		if i == index {
			// Берём только собственный текст, без детей
			txt = strings.TrimSpace(s.Clone().Children().Remove().End().Text())
			return false
		}
		return true
	})
	return txt
}

func parseText(doc *goquery.Document, selector string, index int) string {
	txt := "null"
	doc.Find(selector).EachWithBreak(func(i int, s *goquery.Selection) bool {
		if i == index {
			txt = strings.TrimSpace(s.Text())
			return false
		}
		return true
	})
	return txt
}

func downloadPhotos(doc *goquery.Document, id string) []string {
	var photos []string
	dir := filepath.Join("Result_Crown", id)
	os.MkdirAll(dir, 0755)

	doc.Find(`div[data-ftid="bull-page_bull-gallery_thumbnails"] a`).Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if ok && href != "" {
			resp, err := http.Get(href)
			if err != nil {
				fmt.Println("Error downloading:", href, err)
				return
			}
			defer resp.Body.Close()

			filename := filepath.Join(dir, fmt.Sprintf("%d.jpg", i+1))
			out, err := os.Create(filename)
			if err != nil {
				fmt.Println("Error creating file:", filename, err)
				return
			}
			defer out.Close()

			_, err = io.Copy(out, resp.Body)
			if err != nil {
				fmt.Println("Error saving file:", filename, err)
				return
			}

			photos = append(photos, filename)
		}
	})

	return photos
}

func saveResults(autos []Auto) {
	os.MkdirAll("Result_Crown", 0755)

	file, err := os.Create(filepath.Join("Result_Crown", "Data.csv"))
	if err != nil {
		log.Fatalf("Ошибка создания CSV файла: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{
		"id", "url", "brand", "model", "price", "price_mark", "generation",
		"complectation", "mileage", "no_mileage_rf", "color", "body_type",
		"power", "fuel_type", "engine_volume",
	})

	for _, a := range autos {
		writer.Write([]string{
			a.ID,
			a.URL,
			nullIfEmpty(a.Brand),
			nullIfEmpty(a.Model),
			nullIfEmpty(a.Price),
			nullIfEmpty(a.PriceMark),
			nullIfEmpty(a.Generation),
			nullIfEmpty(a.Complectation),
			nullIfEmpty(a.Mileage),
			nullIfEmpty(a.NoMileageRF),
			nullIfEmpty(a.Color),
			nullIfEmpty(a.BodyType),
			nullIfEmpty(a.Power),
			nullIfEmpty(a.FuelType),
			nullIfEmpty(a.EngineVolume),
		})
	}
}

func nullIfEmpty(s string) string {
	if strings.TrimSpace(s) == "" {
		return "null"
	}
	return s
}

func parseBrandModel(doc *goquery.Document) (brand, model string) {
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

func parsePrice(doc *goquery.Document, selector string) string {
	txt := strings.TrimSpace(doc.Find(selector).First().Text())
	clean := strings.ReplaceAll(txt, "\u00A0", "")
	clean = strings.ReplaceAll(clean, " ", "")
	clean = strings.ReplaceAll(clean, "₽", "")
	if clean == "" {
		return "null"
	}
	return clean
}

func cleanText(s string) string {
	s = strings.ReplaceAll(s, "\u00A0", "")
	s = strings.ReplaceAll(s, "км", "")
	return strings.TrimSpace(s)
}
