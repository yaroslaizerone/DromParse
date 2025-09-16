package save

import (
	"dromCrownParse/internal/models"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SaveResults сохраняет срез автомобилей autos в CSV файл в директории resultDir.
// Заголовки CSV соответствуют полям модели Auto. Пустые строки заменяются на "null".
// Возвращает ошибку, если не удалось создать папку или файл, либо записать данные.
func SaveResults(autos []models.Auto, resultDir string) error {
	// Создание директории для результатов
	if err := os.MkdirAll(resultDir, 0755); err != nil {
		return fmt.Errorf("ошибка создания директории %s: %w", resultDir, err)
	}

	// Создание CSV файла
	filePath := filepath.Join(resultDir, "Data.csv")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("ошибка создания CSV файла %s: %w", filePath, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Запись заголовков
	if err := writer.Write([]string{
		"id", "url", "brand", "model", "price", "price_mark", "generation",
		"complectation", "mileage", "no_mileage_rf", "color", "body_type",
		"power", "fuel_type", "engine_volume",
	}); err != nil {
		return fmt.Errorf("ошибка записи заголовков CSV: %w", err)
	}

	// Запись данных автомобилей
	for _, a := range autos {
		record := []string{
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
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("ошибка записи строки CSV: %w", err)
		}
	}

	return nil
}

// nullIfEmpty возвращает "null", если строка пустая или содержит только пробелы.
// Иначе возвращает исходную строку.
func nullIfEmpty(s string) string {
	if strings.TrimSpace(s) == "" {
		return "null"
	}
	return s
}
