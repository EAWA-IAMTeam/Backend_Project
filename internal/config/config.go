package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DbHost      string
	DbUser      string
	DbPassword  string
	DbName      string
	DbPort      string
	DbSSLMode   string
	AppKey      string
	AppSecret   string
	AccessToken string
}

func LoadConfig() *Config {
	err := godotenv.Load("config/.env")
	if err != nil {
		log.Printf("Error loading .env file: %v\n", err)
	}

	config := &Config{
		DbHost:      getEnv("DB_HOST", "localhost"),
		DbUser:      getEnv("DB_USER", "postgres"),
		DbPassword:  getEnv("DB_PASSWORD", "postgres"),
		DbName:      getEnv("DB_NAME", "postgres"),
		DbPort:      getEnv("DB_PORT", "5432"),
		DbSSLMode:   getEnv("DB_SSLMODE", "disable"),
		AppKey:      getEnv("App_Key", ""),
		AppSecret:   getEnv("App_Secret", ""),
		AccessToken: getEnv("Access_Token", ""),
	}

	// log.Printf("Loaded Config: %+v\n", config)

	return config

}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
