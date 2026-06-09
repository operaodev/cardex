package handler

import (
	"log"
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/operaodev/cardex/internal/products"
	syncsvc "github.com/operaodev/cardex/internal/sync"
)

// SyncHandler expone el servicio de sincronización vía HTTP.
type SyncHandler struct {
	svc     *syncsvc.SyncService
	running atomic.Bool // evita disparar sincronizaciones paralelas
}

// NewSyncHandler crea una nueva instancia del handler de sync.
func NewSyncHandler(svc *syncsvc.SyncService) *SyncHandler {
	return &SyncHandler{svc: svc}
}

// TriggerSync dispara una sincronización asíncrona en segundo plano.
// Retorna HTTP 202 inmediatamente para no bloquear al cliente.
// Query param: tcg (requerido) — por ejemplo "ygo".
//
//	POST /sync/:tcg
func (h *SyncHandler) TriggerSync(c *gin.Context) {
	tcg := products.TCG(c.Param("tcg"))
	if tcg == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe especificar el TCG (e.g. /sync/ygo)"})
		return
	}

	// Evitar sincronizaciones concurrentes
	if !h.running.CompareAndSwap(false, true) {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Ya hay una sincronización en progreso. Intenta de nuevo más tarde.",
		})
		return
	}

	go func() {
		defer h.running.Store(false)
		n, err := h.svc.SyncAll(tcg)
		if err != nil {
			log.Printf("[sync] ERROR durante sincronización de %s: %v", tcg, err)
			return
		}
		log.Printf("[sync] Sincronización de %s completada: %d cartas procesadas", tcg, n)
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Sincronización iniciada en segundo plano",
		"tcg":     tcg,
	})
}

// TriggerSyncByName dispara una sincronización asíncrona en segundo plano para una carta en particular.
// Query param: tcg (requerido), body: name (requerido)
//
//	POST /sync/:tcg/by-name
func (h *SyncHandler) TriggerSyncByName(c *gin.Context) {
	tcg := products.TCG(c.Param("tcg"))
	if tcg == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe especificar el TCG (e.g. /sync/ygo)"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe proporcionar el nombre de la carta en el body"})
		return
	}

	// Evitar sincronizaciones concurrentes
	if !h.running.CompareAndSwap(false, true) {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Ya hay una sincronización en progreso. Intenta de nuevo más tarde.",
		})
		return
	}

	go func() {
		defer h.running.Store(false)
		n, err := h.svc.SyncByName(tcg, req.Name)
		if err != nil {
			log.Printf("[sync] ERROR durante sincronización por nombre (%s) de %s: %v", req.Name, tcg, err)
			return
		}
		log.Printf("[sync] Sincronización por nombre de %s completada: %d cartas procesadas", tcg, n)
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Sincronización por nombre iniciada en segundo plano",
		"tcg":     tcg,
		"name":    req.Name,
	})
}

// SyncStatus devuelve si hay una sincronización en progreso.
//
//	GET /sync/status
func (h *SyncHandler) SyncStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"running": h.running.Load(),
	})
}
