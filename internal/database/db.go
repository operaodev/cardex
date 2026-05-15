package database

import (
	"fmt"
	"log"
	"os"

	"github.com/operaodev/cardex/internal/cards"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_NAME", "postgres"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_SSLMODE", "require"),
		getEnv("DB_TIMEZONE", "UTC"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})

	if err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}

	// AutoMigrate crea o actualiza la tabla 'cards' con todas las columnas,
	// JSONBs, índices compuestos y GIN definidos en el modelo.
	if err = db.AutoMigrate(&cards.Card{}); err != nil {
		log.Fatalf("Error en automigración: %v", err)
	}

	log.Println("Conectado a PostgreSQL y base de datos migrada con éxito")
	DB = db
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
