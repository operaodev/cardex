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

// GetFromProvider handles GET /cards/provider/:provider/:id
// It fetches all localised versions of a card by its YGOPro numeric ID.
func (h *Handler) GetFromProvider(c *gin.Context) {
	provider := c.Param("provider")
	id := c.Param("id")

	results, err := h.svc.SearchByID(provider, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}
