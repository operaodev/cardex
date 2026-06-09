package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	custompacks "github.com/operaodev/cardex/internal/custom_packs"
)

type WishlistHandler struct {
	service custompacks.Service
}

func NewWishlistHandler(s custompacks.Service) *WishlistHandler {
	return &WishlistHandler{service: s}
}

// UpsertRequest es el cuerpo de la petición para agregar/quitar items.
type UpsertRequest struct {
	ProductID uint64 `json:"product_id" binding:"required"`
	Delta     int    `json:"delta"     binding:"required"`
}

// GetMyWishlist obtiene la wishlist del usuario autenticado.
// GET /wishlist
func (h *WishlistHandler) GetMyWishlist(c *gin.Context) {
	userID, _ := c.Get("userID")
	items, err := h.service.GetByUserID(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// Upsert agrega o modifica un item de la wishlist.
// POST /wishlist
func (h *WishlistHandler) Upsert(c *gin.Context) {
	var req UpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidJSONBody})
		return
	}

	userID, _ := c.Get("userID")
	item, err := h.service.Upsert(userID.(string), req.ProductID, req.Delta)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if item == nil {
		c.JSON(http.StatusOK, gin.H{"message": "item eliminado por cantidad <= 0"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// Delete elimina un item de la wishlist.
// DELETE /wishlist/:product_id
func (h *WishlistHandler) Delete(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id inválido"})
		return
	}

	userID, _ := c.Get("userID")
	if err := h.service.Delete(userID.(string), productID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "item eliminado"})
}
