package sync

import (
	"fmt"
	"log"

	"github.com/operaodev/cardex/internal/items"
	"github.com/operaodev/cardex/internal/providers"
)

// SyncService orquesta la sincronización de cartas externas hacia la DB local.
type SyncService struct {
	providersSvc *providers.Service
	itemsRepo    items.Repository
}

func NewSyncService(providersSvc *providers.Service, itemsRepo items.Repository) *SyncService {
	return &SyncService{
		providersSvc: providersSvc,
		itemsRepo:    itemsRepo,
	}
}

// SyncAll obtiene todas las cartas del proveedor indicado y las persiste en la DB.
// Devuelve el número de cartas procesadas (nuevas + actualizadas) y un error si falla.
func (s *SyncService) SyncAll(tcg items.TCG) (int, error) {
	log.Printf("[sync] Iniciando sincronización completa para TCG=%s", tcg)

	results, err := s.providersSvc.FetchCards(tcg)
	if err != nil {
		return 0, fmt.Errorf("error obteniendo cartas del proveedor: %w", err)
	}

	if len(results) == 0 {
		log.Printf("[sync] No se encontraron cartas para TCG=%s", tcg)
		return 0, nil
	}

	log.Printf("[sync] %d cartas encontradas. Persistiendo...", len(results))

	upserted, err := s.itemsRepo.Upsert(results)
	if err != nil {
		return 0, fmt.Errorf("error persistiendo cartas en la DB: %w", err)
	}

	log.Printf("[sync] Sincronización completada: %d cartas procesadas", upserted)
	return upserted, nil
}

// SyncByName obtiene cartas del proveedor por nombre y las persiste en la DB.
func (s *SyncService) SyncByName(tcg items.TCG, name string) (int, error) {
	log.Printf("[sync] Iniciando sincronización por nombre para TCG=%s, Name=%s", tcg, name)

	results, err := s.providersSvc.FetchCardsByName(tcg, name)
	if err != nil {
		return 0, fmt.Errorf("error obteniendo cartas del proveedor por nombre: %w", err)
	}

	if len(results) == 0 {
		log.Printf("[sync] No se encontraron cartas para TCG=%s, Name=%s", tcg, name)
		return 0, nil
	}

	log.Printf("[sync] %d cartas encontradas. Persistiendo...", len(results))

	upserted, err := s.itemsRepo.Upsert(results)
	if err != nil {
		return 0, fmt.Errorf("error persistiendo cartas en la DB: %w", err)
	}

	log.Printf("[sync] Sincronización por nombre completada: %d cartas procesadas", upserted)
	return upserted, nil
}
