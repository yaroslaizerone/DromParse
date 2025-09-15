package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config хранит все настройки приложения
type Config struct {
	AppName         string
	AppEnv          string
	AppDebug        bool
	ResultDir       string
	MaxWorkers      int
	RequestTimeout  int
	RequestDelayMin int
	RequestDelayMax int
	Cities          []string
	Proxy           string
}

// AppConfig Глобальная переменная конфигурации
var AppConfig Config

// LoadConfig загружает настройки из .env
func LoadConfig() {
	// загружаем .env (если есть)
	_ = godotenv.Load()

	AppConfig = Config{
		AppName:         getEnv("APP_NAME", "DromParser"),
		AppEnv:          getEnv("APP_ENV", "local"),
		AppDebug:        getEnvBool("APP_DEBUG", true),
		ResultDir:       getEnv("RESULT_DIR", "Result_Crown"),
		MaxWorkers:      getEnvInt("MAX_WORKERS", 5),
		RequestTimeout:  getEnvInt("REQUEST_TIMEOUT", 15),
		RequestDelayMin: getEnvInt("REQUEST_DELAY_MIN", 1),
		RequestDelayMax: getEnvInt("REQUEST_DELAY_MAX", 3),
		Cities:          getEnvSlice("CITIES", []string{"https://vladivostok.drom.ru/auto/all/", "https://ussuriisk.drom.ru/auto/all/"}),
		Proxy:           getEnv("PROXY", ""),
	}
}

// --- вспомогательные функции ---
func getEnv(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func getEnvBool(key string, defaultVal bool) bool {
	val := strings.ToLower(os.Getenv(key))
	if val == "true" || val == "1" {
		return true
	} else if val == "false" || val == "0" {
		return false
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Ошибка конвертации %s=%s, используем default=%d", key, val, defaultVal)
		return defaultVal
	}
	return n
}

func getEnvSlice(key string, defaultVal []string) []string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	parts := strings.Split(val, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
