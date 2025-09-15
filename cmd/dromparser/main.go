package main

import (
	"dromCrownParse/internal/fetch"
	"dromCrownParse/internal/utils"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"sync"

	"dromCrownParse/internal/config"
	"dromCrownParse/internal/models"
	"dromCrownParse/internal/save"
)

func main() {
	// Загружаем конфиг из .env
	config.LoadConfig()

	cities := config.AppConfig.Cities

	// Парсим фильтры из строки
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
	var autos []models.Auto
	maxWorkers := config.AppConfig.MaxWorkers

	for _, base := range cities {
		fmt.Printf("\n=== Сканирую город: %s ===\n", base)
		firstURL := base + "?" + params.Encode()

		doc, err := fetch.FetchDocument(firstURL)
		if err != nil {
			log.Printf("Ошибка первого запроса: %v", err)
			continue
		}

		maxPage := utils.GetMaxPage(doc)
		totalCount := utils.GetTotalCount(doc, numRe)

		fmt.Printf("Всего страниц: %d, объявлений: %d\n", maxPage, totalCount)

		collected := 0
		pagesCh := make(chan int, maxPage)
		resultsCh := make(chan models.Auto, totalCount)
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
					doc, err := fetch.FetchDocument(pageURL)
					if err != nil {
						continue
					}
					fetch.ProcessPage(doc, base, idRe, &seen, &mu, &collected, totalCount, resultsCh)
				}
			}()
		}

		wg.Wait()
		close(resultsCh)

		for a := range resultsCh {
			autos = append(autos, a)
		}
	}

	save.SaveResults(autos, config.AppConfig.ResultDir)
	fmt.Printf("\nГотово: %d автомобилей\n", len(autos))
}
