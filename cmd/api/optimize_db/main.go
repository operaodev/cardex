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

	log.Println("Habilitando extensión pg_trgm...")
	database.DB.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm;")

	log.Println("Creando índice GIN para búsquedas ultra rápidas por nombre...")
	// Este índice permite que ILIKE %termino% use un índice en lugar de Sequential Scan
	database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_cards_name_trgm ON cards USING gin (name gin_trgm_ops);")
	
	log.Println("Creando índice para arquetipos...")
	database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_cards_archetype_trgm ON cards USING gin (archetype gin_trgm_ops);")

	log.Println("Optimizando tablas (ANALYZE)...")
	database.DB.Exec("ANALYZE cards;")

	log.Println("¡Base de Datos optimizada con éxito!")
}
