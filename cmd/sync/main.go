package main

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/operaodev/cardex/internal/cards"
	"github.com/operaodev/cardex/internal/database"
	"github.com/operaodev/cardex/internal/search"
	searchproviders "github.com/operaodev/cardex/internal/search/providers"
	syncsvc "github.com/operaodev/cardex/internal/sync"
)

func main() {
	// Flags CLI
	tcg := flag.String("tcg", "ygo", "TCG a sincronizar (ygo, mtg, pkm)")
	envFile := flag.String("env", ".env", "Ruta al archivo .env")
	flag.Parse()

	// Cargar variables de entorno
	if err := godotenv.Load(*envFile); err != nil {
		log.Printf("[sync] Advertencia: no se pudo cargar %s: %v", *envFile, err)
	}

	// Inicializar base de datos
	database.Connect()

	// Capas de la aplicación
	repo := cards.NewRepository(database.DB)
	ygoProv := searchproviders.NewYGOProvider()
	searchSvc := search.NewService(ygoProv)
	svc := syncsvc.NewSyncService(searchSvc, repo)

	log.Printf("[sync] Iniciando sincronización manual para TCG=%s", *tcg)

	n, err := svc.SyncAll(*tcg)
	if err != nil {
		log.Printf("[sync] ERROR: %v", err)
		os.Exit(1)
	}

	log.Printf("[sync] Completado: %d cartas procesadas", n)
}
