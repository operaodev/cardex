package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/operaodev/cardex/internal/items"
	"github.com/operaodev/cardex/internal/providers"
)

type ProviderHandler struct {
	svc *providers.Service
}

func NewProviderHandler(svc *providers.Service) *ProviderHandler {
	return &ProviderHandler{
		svc: svc,
	}
}

func (h *ProviderHandler) FetchCards(c *gin.Context) {
	provider := c.Param("provider")

	result, err := h.svc.FetchCards(items.TCG(provider))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *ProviderHandler) FetchCardsByName(c *gin.Context) {
	provider := c.Param("provider")
	name := c.Param("name")

	result, err := h.svc.FetchCardsByName(items.TCG(provider), name)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		return
	}

	if result == nil {
		result = []items.Item{}
	}

	c.JSON(http.StatusOK, result)
}
