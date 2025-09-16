package photos

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/PuerkitoBio/goquery"
)

// DownloadPhotos загружает все фотографии автомобиля с документа goquery и сохраняет их в папку Result_Crown/<id>.
// Возвращает срез локальных путей к сохранённым файлам.
// Любые ошибки при создании папки, загрузке или сохранении файлов логируются, но не прерывают выполнение функции.
func DownloadPhotos(doc *goquery.Document, id string) []string {
	var photos []string
	dir := filepath.Join("Result_Crown", id)

	// Создание папки для фотографий
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Println("Ошибка создания папки:", dir, err)
		return photos
	}

	// Поиск ссылок на фотографии
	doc.Find(`div[data-ftid="bull-page_bull-gallery_thumbnails"] a`).Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok || href == "" {
			return
		}

		// Загрузка фотографии
		resp, err := http.Get(href)
		if err != nil {
			fmt.Println("Ошибка загрузки:", href, err)
			return
		}
		defer resp.Body.Close()

		// Создание файла
		filename := filepath.Join(dir, fmt.Sprintf("%d.jpg", i+1))
		out, err := os.Create(filename)
		if err != nil {
			fmt.Println("Ошибка создания файла:", filename, err)
			return
		}
		defer out.Close()

		// Сохранение данных в файл
		if _, err := io.Copy(out, resp.Body); err != nil {
			fmt.Println("Ошибка сохранения файла:", filename, err)
			return
		}

		photos = append(photos, filename)
	})

	return photos
}
