package downloader

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"dromCrownParse/config"
	"github.com/PuerkitoBio/goquery"
)

// seenPhotos используется для отслеживания уже загруженных файлов
var seenPhotos sync.Map

// PhotoJob описывает задачу на скачивание одного фото
type PhotoJob struct {
	URL string // URL изображения
	Dir string // Папка для сохранения
	Idx int    // Индекс изображения (для имени файла)
}

// DownloadPhotosParallel скачивает все фотографии с объявления параллельно.
// Принимает http.Client для запросов, goquery.Document с HTML страницы,
// id авто для создания папки и конфиг с настройками.
// Возвращает список локальных путей к скачанным фотографиям.
func DownloadPhotosParallel(client *http.Client, doc *goquery.Document, id string, cfg *config.Config) []string {
	var photos []string
	dir := filepath.Join(cfg.ResultDir, id)
	os.MkdirAll(dir, 0755)

	var jobs []PhotoJob
	doc.Find(`div[data-ftid="bull-page_bull-gallery_thumbnails"] a`).Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok && href != "" {
			jobs = append(jobs, PhotoJob{URL: href, Dir: dir, Idx: i})
		}
	})

	jobCh := make(chan PhotoJob, len(jobs))
	var wg sync.WaitGroup

	for w := 0; w < cfg.MaxWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobCh {
				filename := filepath.Join(job.Dir, fmt.Sprintf("%d.jpg", job.Idx+1))
				if _, exists := seenPhotos.LoadOrStore(filename, true); exists {
					continue
				}
				if err := downloadFile(client, job.URL, filename); err != nil {
					fmt.Println("Ошибка скачивания:", job.URL, err)
				} else {
					photos = append(photos, filename)
				}
				time.Sleep(time.Duration(cfg.DelayMin+rand.Intn(cfg.DelayMax-cfg.DelayMin+1)) * time.Second)
			}
		}()
	}

	for _, job := range jobs {
		jobCh <- job
	}
	close(jobCh)
	wg.Wait()
	return photos
}

// downloadFile скачивает файл с указанного URL и сохраняет его по указанному пути.
// Используется в DownloadPhotosParallel. Возвращает ошибку при неудаче.
func downloadFile(client *http.Client, url, filepath string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}
