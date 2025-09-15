package tests

import (
	"dromCrownParse/internal/utils"
	"regexp"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestGetMaxPage(t *testing.T) {
	html := `
		<html>
			<body>
				<a data-ftid="component_pagination-item">1</a>
				<a data-ftid="component_pagination-item">2</a>
				<a data-ftid="component_pagination-item">5</a>
			</body>
		</html>
	`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Ошибка создания документа: %v", err)
	}

	maxPage := utils.GetMaxPage(doc)
	if maxPage != 5 {
		t.Errorf("Ожидалось maxPage=5, получили %d", maxPage)
	}
}

func TestGetTotalCount(t *testing.T) {
	html := `
		<html>
			<body>
				<a data-ftid="bulls-list_bulls-tab">
					<span class="_3ynq47p">Всего: 42</span>
				</a>
				<div id="tabs">
					<span>99</span>
				</div>
			</body>
		</html>
	`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Ошибка создания документа: %v", err)
	}

	numRe := regexp.MustCompile(`(\d+)`)
	totalCount := utils.GetTotalCount(doc, numRe)
	if totalCount != 42 {
		t.Errorf("Ожидалось totalCount=42, получили %d", totalCount)
	}

	// Тест с пустым bulls-tab, чтобы проверить fallback на div#tabs
	htmlFallback := `
		<html>
			<body>
				<a data-ftid="bulls-list_bulls-tab">
					<span class="_3ynq47p">Всего: 0</span>
				</a>
				<div id="tabs">
					<span>99</span>
				</div>
			</body>
		</html>
	`
	docFallback, _ := goquery.NewDocumentFromReader(strings.NewReader(htmlFallback))
	totalCountFallback := utils.GetTotalCount(docFallback, numRe)
	if totalCountFallback != 99 {
		t.Errorf("Ожидалось totalCountFallback=99, получили %d", totalCountFallback)
	}
}
