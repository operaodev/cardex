package main

import (
	"log"
	"github.com/joho/godotenv"
	"github.com/operaodev/cardex/internal/database"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Advertencia: no se pudo cargar .env: %v", err)
	}
	database.Connect()
	log.Println("Ejecutando DROP INDEX para que GORM regenere la nueva versión de idx_card_identity...")
	database.DB.Exec("DROP INDEX IF EXISTS idx_card_identity;")
	log.Println("Índice eliminado. Ejecutando automigración para recrearlo...")
	database.Connect() // Re-conecta para disparar AutoMigrate de nuevo
	log.Println("Listo.")
}
