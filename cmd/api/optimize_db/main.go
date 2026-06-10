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

	log.Println("Creando índice GIN para búsquedas ultra rápidas por nombre en cards...")
	// Este índice permite que ILIKE %termino% use un índice en lugar de Sequential Scan
	database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_cards_name_trgm ON cards USING gin (name gin_trgm_ops);")

	log.Println("Creando índice para arquetipos en cards...")
	database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_cards_archetype_trgm ON cards USING gin (archetype gin_trgm_ops);")

	log.Println("Creando índices GIN para búsquedas en products...")
	database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_products_name_trgm ON products USING gin (name gin_trgm_ops);")
	database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_products_code_trgm ON products USING gin (code gin_trgm_ops);")
	database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_products_archetype_trgm ON products USING gin (archetype gin_trgm_ops);")
	database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_products_external_id_trgm ON products USING gin (external_id gin_trgm_ops);")

	log.Println("Optimizando tablas (ANALYZE)...")
	database.DB.Exec("ANALYZE cards;")
	database.DB.Exec("ANALYZE products;")

	log.Println("¡Base de Datos optimizada con éxito!")
}
