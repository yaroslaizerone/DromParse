package parser

import (
	"dromCrownParse/internal/models"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func ParseText(doc *goquery.Document, selector string, index int) string {
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

func ParsePrice(doc *goquery.Document, selector string) string {
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

func ParseTableRow(a *models.Auto, th, td string) {
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
}
