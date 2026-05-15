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
	log.Println("Vaciando la tabla de cartas para eliminar duplicados...")
	database.DB.Exec("TRUNCATE TABLE cards RESTART IDENTITY CASCADE;")
	log.Println("Tabla vaciada. Ejecutando automigración para asegurar que el índice único se cree...")
	database.Connect() // Re-conecta para disparar AutoMigrate
	log.Println("¡Listo! La base de datos está limpia y con el índice único aplicado.")
}
