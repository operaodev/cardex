package search

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{svc: s}
}

// SearchByIDInProvider handles GET /cards/search/:provider/:id
// Busca una carta por su ID en el proveedor.
func (h *Handler) SearchByIDInProvider(c *gin.Context) {
	provider := c.Param("provider")
	id := c.Param("id")

	result, err := h.svc.SearchByID(provider, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// SearchByNamesInProvider handles GET /cards/search/:provider?name=Kuriboh
// Busca cartas por sus nombres en el proveedor.
func (h *Handler) SearchByNamesInProvider(c *gin.Context) {
	provider := c.Param("provider")
	name := c.Query("name")

	results, err := h.svc.SearchByNames(provider, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

