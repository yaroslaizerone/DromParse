package save

import (
	"dromCrownParse/internal/models"
	"encoding/csv"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func SaveResults(autos []models.Auto, resultDir string) {
	os.MkdirAll(resultDir, 0755)

	file, err := os.Create(filepath.Join(resultDir, "Data.csv"))
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
