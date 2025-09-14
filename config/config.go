package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config хранит конфигурационные параметры приложения
type Config struct {
	ResultDir      string   // Директория для сохранения результатов
	MaxWorkers     int      // Максимальное количество воркеров для параллельной обработки
	RequestTimeout int      // Таймаут HTTP-запросов в секундах
	DelayMin       int      // Минимальная задержка между запросами в секундах
	DelayMax       int      // Максимальная задержка между запросами в секундах
	Cities         []string // Список URL городов для сканирования
	Filters        string   // Параметры фильтров для запросов
}

// LoadConfig загружает конфигурацию из .env файла или возвращает значения по умолчанию.
// Если .env файл не найден, выводится предупреждение.
func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println(".env файл не найден, используются значения по умолчанию")
	}

	return &Config{
		ResultDir:      getEnv("RESULT_DIR", "Result_Crown"),
		MaxWorkers:     getEnvInt("MAX_WORKERS", 5),
		RequestTimeout: getEnvInt("REQUEST_TIMEOUT", 15),
		DelayMin:       getEnvInt("REQUEST_DELAY_MIN", 1),
		DelayMax:       getEnvInt("REQUEST_DELAY_MAX", 3),
		Cities:         strings.Split(getEnv("CITIES", "https://vladivostok.drom.ru/auto/all/,https://ussuriisk.drom.ru/auto/all/"), ","),
		Filters:        getEnv("FILTERS", "multiselect[]=9_4_15_all&multiselect[]=9_4_16_all&ph=1&pts=2&damaged=2&unsold=1&whereabouts[]=0"),
	}
}

// getEnv возвращает значение переменной окружения по ключу.
// Если переменная не задана, возвращается значение fallback.
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// getEnvInt возвращает значение переменной окружения в виде int.
// Если переменная не задана или не может быть преобразована в число, возвращается fallback.
func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return fallback
}
