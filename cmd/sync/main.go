package main

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/operaodev/cardex/internal/database"
	"github.com/operaodev/cardex/internal/products"
	"github.com/operaodev/cardex/internal/providers"
	syncsvc "github.com/operaodev/cardex/internal/sync"
)

func main() {
	// Flags CLI
	tcg := flag.String("tcg", "ygo", "TCG a sincronizar (ygo, mtg, pkm)")
	name := flag.String("name", "", "Nombre de la carta para sincronizar una específica (opcional)")
	envFile := flag.String("env", ".env", "Ruta al archivo .env")
	flag.Parse()

	// Cargar variables de entorno
	if err := godotenv.Load(*envFile); err != nil {
		log.Printf("[sync] Advertencia: no se pudo cargar %s: %v", *envFile, err)
	}

	// Inicializar base de datos
	database.Connect()

	// Capas de la aplicación
	itemsRepo := products.NewRepository(database.DB)
	ygoProv := providers.NewYGOProvider()
	providerSvc := providers.NewService(ygoProv)
	svc := syncsvc.NewSyncService(providerSvc, itemsRepo)

	var n int
	var err error

	tcgEnum := products.TCG(*tcg)

	if *name != "" {
		log.Printf("[sync] Iniciando sincronización manual por nombre para TCG=%s, Name=%s", *tcg, *name)
		n, err = svc.SyncByName(tcgEnum, *name)
	} else {
		log.Printf("[sync] Iniciando sincronización manual para TCG=%s", *tcg)
		n, err = svc.SyncAll(tcgEnum)
	}

	if err != nil {
		log.Printf("[sync] ERROR: %v", err)
		os.Exit(1)
	}

	log.Printf("[sync] Completado: %d cartas procesadas", n)
}
