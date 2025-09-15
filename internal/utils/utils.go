package utils

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func GetMaxPage(doc *goquery.Document) int {
	maxPage := 1
	doc.Find(`a[data-ftid="component_pagination-item"]`).Each(func(i int, s *goquery.Selection) {
		txt := strings.TrimSpace(s.Text())
		if n, err := strconv.Atoi(txt); err == nil && n > maxPage {
			maxPage = n
		}
	})
	return maxPage
}

func GetTotalCount(doc *goquery.Document, numRe *regexp.Regexp) int {
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
	return totalCount
}
