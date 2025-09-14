package dromparser

import (
	"dromCrownParse/config"
	"dromCrownParse/models"
	"dromCrownParse/parser"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func main() {
	cfg := config.LoadConfig()
	client := &http.Client{
		Timeout: time.Duration(cfg.RequestTimeout) * time.Second,
	}

	var autos []models.Auto
	var mu sync.Mutex

	for _, base := range cfg.Cities {
		fmt.Printf("\n=== Сканирую город: %s ===\n", base)

		firstURL := base + "?" + cfg.Filters
		doc, err := parser.FetchDocument(client, firstURL, cfg)
		if err != nil {
			log.Printf("Ошибка при загрузке страницы: %v", err)
			continue
		}

		maxPage, totalCount := parser.ExtractPaginationInfo(doc)
		fmt.Printf("Всего страниц: %d, объявлений: %d\n", maxPage, totalCount)

		collected := 0
		pagesCh := make(chan int, maxPage)
		resultsCh := make(chan models.Auto, totalCount)
		var wg sync.WaitGroup

		for page := 1; page <= maxPage; page++ {
			pagesCh <- page
		}
		close(pagesCh)

		for w := 0; w < cfg.MaxWorkers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for page := range pagesCh {
					if collected >= totalCount && totalCount > 0 {
						return
					}
					pageURL := fmt.Sprintf("%spage%d/?%s", base, page, cfg.Filters)
					pageDoc, err := parser.FetchDocument(client, pageURL, cfg)
					if err != nil {
						continue
					}

					parser.CollectAutosFromPage(pageDoc, &autos, &mu, resultsCh, &collected, totalCount)
				}
			}()
		}

		wg.Wait()
		close(resultsCh)

		for a := range resultsCh {
			autos = append(autos, a)
		}
	}

	if err := SaveResults(cfg.ResultDir, autos); err != nil {
		log.Fatalf("Ошибка сохранения результатов: %v", err)
	}

	fmt.Printf("\nГотово: %d автомобилей\n", len(autos))
}

// SaveResults сохраняет список автомобилей в CSV-файл в указанной директории.
//
// Параметры:
//   - dir: Путь к директории, в которой будет создан файл "Data.csv". Директория создаётся, если её нет.
//   - autos: Срез объектов models.Auto, каждый из которых содержит информацию об автомобиле.
//
// Формат CSV:
//
//	id, url, brand, model, price, price_mark, generation, complectation, mileage, no_mileage_rf,
//	color, body_type, power, fuel_type, engine_volume
//
// Для каждого поля авто, если значение пустое, будет записано "null".
//
// Возвращает:
//   - error: Ошибка создания файла или записи, если она возникла. В случае успешного сохранения возвращается nil.
func SaveResults(dir string, autos []models.Auto) error {
	os.MkdirAll(dir, 0755)

	file, err := os.Create(filepath.Join(dir, "Data.csv"))
	if err != nil {
		return err
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

	return nil
}

// nullIfEmpty проверяет строку и возвращает "null", если она пустая или состоит только из пробелов.
//
// Параметры:
//   - s: Исходная строка для проверки.
//
// Возвращает:
//   - string: Исходную строку без изменений, если она не пустая, или "null", если строка пустая или содержит только пробелы.
func nullIfEmpty(s string) string {
	if strings.TrimSpace(s) == "" {
		return "null"
	}
	return s
}
