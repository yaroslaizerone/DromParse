package tests

import (
	"strings"
	"testing"

	"dromCrownParse/parser"
	"github.com/PuerkitoBio/goquery"
)

const sampleHTML = `<a data-ftid="header_breadcrumb_link">
<span class="_1lj8ai62">Toyota</span></a>
<a data-ftid="header_breadcrumb_link">
<span class="_1lj8ai62">Camry</span></a>`

func TestParseBrandModel(t *testing.T) {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(sampleHTML))
	brand, model := parser.ParseBrandModel(doc)
	if brand != "Toyota" || model != "Camry" {
		t.Errorf("Ожидается Toyota Camry, получили %s %s", brand, model)
	}
}

const missingHTML = `<a data-ftid="header_breadcrumb_link">
<span class="_1lj8ai62"></span></a>`

func TestParseBrandModel_Empty(t *testing.T) {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(missingHTML))
	brand, model := parser.ParseBrandModel(doc)
	if brand != "null" || model != "null" {
		t.Errorf("Ожидается null null, получили %s %s", brand, model)
	}
}

const paginationHTML = `
<div>
  <a data-ftid="component_pagination-item">1</a>
  <a data-ftid="component_pagination-item">2</a>
  <a data-ftid="component_pagination-item">3</a>
</div>
<a data-ftid="bulls-list_bulls-tab">
  <span class="_3ynq47p">Total 42</span>
</a>
`

func TestExtractPaginationInfo(t *testing.T) {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(paginationHTML))
	maxPage, totalCount := parser.ExtractPaginationInfo(doc)
	if maxPage != 3 {
		t.Errorf("Ожидается maxPage = 3, получили %d", maxPage)
	}
	if totalCount != 42 {
		t.Errorf("Ожидается totalCount = 42, получили %d", totalCount)
	}
}
