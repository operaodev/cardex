package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/operaodev/cardex/internal/cards"
)

// CardsHandler expone las funciones del servicio a través de HTTP.
type CardsHandler struct {
	service cards.Service
}

// NewCardsHandler crea una nueva instancia del Handler inyectando el servicio.
func NewCardsHandler(s cards.Service) *CardsHandler {
	return &CardsHandler{
		service: s,
	}
}

// GetByID maneja las peticiones para obtener una carta por ID.
func (h *CardsHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe proporcionar el ID de la carta"})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	card, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, card)
}

// GetByName maneja las peticiones de autocompletado/recomendaciones por nombre.
// Query params: name (requerido), tcg (opcional), lang (opcional)
func (h *CardsHandler) GetSuggestions(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe proporcionar el parámetro 'name'"})
		return
	}

	tcg := cards.TCG(c.Query("tcg"))
	lang := cards.LangCode(c.Query("lang"))

	results, err := h.service.GetSuggestions(tcg, lang, name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetByFilters maneja las peticiones del catálogo con filtros y paginación.
// Query params: tcg, lang, name, type, archetype, subtype, set_code, rarity, page, limit
func (h *CardsHandler) GetCatalog(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	filters := cards.CatalogFilters{
		Name:      c.Query("name"),
		TCG:       cards.TCG(c.Query("tcg")),
		Lang:      cards.LangCode(c.Query("lang")),
		Type:      c.Query("type"),
		Archetype: c.Query("archetype"),
		Subtype:   c.Query("subtype"),
		SetCode:   c.Query("set_code"),
		Rarity:    cards.Rarity(c.Query("rarity")),
		Page:      page,
		Limit:     limit,
	}

	result, err := h.service.GetCatalog(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
