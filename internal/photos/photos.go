package photos

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/PuerkitoBio/goquery"
)

func DownloadPhotos(doc *goquery.Document, id string) []string {
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
