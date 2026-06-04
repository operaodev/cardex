package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/operaodev/cardex/internal/items"
)

// ItemsHandler expone las funciones del servicio de ítems a través de HTTP.
type ItemsHandler struct {
	service items.Service
}

// NewItemsHandler crea una nueva instancia del Handler inyectando el servicio.
func NewItemsHandler(s items.Service) *ItemsHandler {
	return &ItemsHandler{service: s}
}

// GetByID maneja GET /items/:id
func (h *ItemsHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe proporcionar el ID del ítem"})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	item, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// GetRandomNames maneja GET /items/random/:count
func (h *ItemsHandler) GetRandomNames(c *gin.Context) {
	countStr := c.Param("count")
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 {
		count = 10
	}

	names, err := h.service.GetRandomNames(count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"names": names})
}

// FindSuggestions maneja POST /items/suggestions
// Query params: input, tcg (opcional), lang (opcional)
func (h *ItemsHandler) FindSuggestions(c *gin.Context) {
	var input items.SuggestionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidJSONBody})
		return
	}
	
	if input.Input == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe proporcionar el parámetro 'input'"})
		return
	}

	results, err := h.service.GetSuggestions(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if results == nil {
		results = []items.SuggestionDTO{}
	}

	c.JSON(http.StatusOK, results)
}
