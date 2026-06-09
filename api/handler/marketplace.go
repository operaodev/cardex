package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/operaodev/cardex/internal/marketplace"
)

type MarketplaceHandler struct {
	service marketplace.Service
}

func NewMarketplaceHandler(s marketplace.Service) *MarketplaceHandler {
	return &MarketplaceHandler{service: s}
}

// GetPrices maneja GET /marketplace/analysis/:id
func (h *MarketplaceHandler) GetPrices(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe proporcionar el ID del producto"})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	analysis, err := h.service.GetPrices(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// GetOffers maneja GET /marketplace/offers/:id
// Query params: for_sale (true/false), for_trade (true/false), sort (asc/desc), page, limit
func (h *MarketplaceHandler) GetOffers(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe proporcionar el ID del producto"})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	input := marketplace.OffersInput{
		ProductID: id,
		Page:      1,
		Limit:     20,
	}

	if v := c.Query("for_sale"); v != "" {
		val := v == "true"
		input.ForSale = &val
	}
	if v := c.Query("for_trade"); v != "" {
		val := v == "true"
		input.ForTrade = &val
	}
	if v := c.Query("has_stock"); v != "" {
		val := v == "true"
		input.HasStock = &val
	}
	if v := c.Query("sort"); v == "desc" {
		input.SortDesc = true
	}
	if v := c.Query("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			input.Page = p
		}
	}
	if v := c.Query("limit"); v != "" {
		if l, err := strconv.Atoi(v); err == nil && l > 0 {
			input.Limit = l
		}
	}

	page, err := h.service.GetOffers(input)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, page)
}
