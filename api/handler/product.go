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

// FindSuggestions maneja POST /items/suggestions
// Query params: input, tcg (opcional), lang (opcional)
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
