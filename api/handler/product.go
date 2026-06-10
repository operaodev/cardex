package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/operaodev/cardex/internal/products"
)

// ProductsHandler expone las funciones del servicio de products a través de HTTP.
type ProductsHandler struct {
	service products.Service
}

// NewProductsHandler crea una nueva instancia del Handler inyectando el servicio.
func NewProductsHandler(s products.Service) *ProductsHandler {
	return &ProductsHandler{service: s}
}

// GetByID maneja GET /items/:id
func (h *ProductsHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe proporcionar el ID del product"})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	product, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

// GetRandomNames maneja GET /items/random/:count
func (h *ProductsHandler) GetRandomNames(c *gin.Context) {
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

// FindSuggestions maneja POST /products/suggestions
// Si la petición viene de un usuario autenticado, incluye stock y wishlist.
func (h *ProductsHandler) FindSuggestions(c *gin.Context) {
	var input products.SuggestionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidJSONBody})
		return
	}

	if input.Input == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe proporcionar el parámetro 'input'"})
		return
	}

	userID, exists := c.Get("userID")
	if exists {
		results, err := h.service.GetSuggestionsByUser(userID.(string), input)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if results == nil {
			results = []products.SuggestionDTO{}
		}
		c.JSON(http.StatusOK, results)
		return
	}

	results, err := h.service.GetSuggestions(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if results == nil {
		results = []products.SuggestionDTO{}
	}

	c.JSON(http.StatusOK, results)
}

// FindSuggestionsSimple maneja POST /products/suggestions/simple
// Versión rápida sin auth ni JOINs a stocks/wishlist.
func (h *ProductsHandler) FindSuggestionsSimple(c *gin.Context) {
	var input products.SuggestionInput
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
		results = []products.SuggestionDTO{}
	}

	c.JSON(http.StatusOK, results)
}

// GetRelatedCards maneja POST /products/related
func (h *ProductsHandler) GetRelatedCards(c *gin.Context) {
	var input products.RelatedCardsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	result, err := h.service.GetRelatedCards(input)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetCardsBySet maneja POST /products/set
func (h *ProductsHandler) GetCardsBySet(c *gin.Context) {
	var input products.GetCardsBySetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cuerpo de la petición inválido"})
		return
	}

	if input.SetExternalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "set_external_id requerido"})
		return
	}

	if input.Lang == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lang requerido"})
		return
	}

	cards, err := h.service.GetCardsBySet(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if cards == nil {
		cards = []products.Product{}
	}

	c.JSON(http.StatusOK, cards)
}
