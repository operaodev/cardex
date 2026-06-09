package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/operaodev/cardex/internal/stock"
)

type StockHandler struct {
	service stock.Service
}

func NewStockHandler(s stock.Service) *StockHandler {
	return &StockHandler{service: s}
}

// Create crea un nuevo stock con log de tipo "add".
// POST /stock
func (h *StockHandler) Create(c *gin.Context) {
	var input stock.CreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cuerpo de la petición inválido"})
		return
	}

	userID, _ := c.Get("userID")
	input.UserID = userID.(string)

	s, err := h.service.Create(input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, s)
}

// Restock añade cantidad al stock existente.
// POST /stock/restock
func (h *StockHandler) Restock(c *gin.Context) {
	var input stock.QuantityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cuerpo de la petición inválido"})
		return
	}

	s, err := h.service.Restock(input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, s)
}

// Return registra una devolución de stock.
// POST /stock/return
func (h *StockHandler) Return(c *gin.Context) {
	var input stock.QuantityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cuerpo de la petición inválido"})
		return
	}

	s, err := h.service.Return(input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, s)
}

// Sale registra una venta de stock.
// POST /stock/sale
func (h *StockHandler) Sale(c *gin.Context) {
	var input stock.DecreaseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cuerpo de la petición inválido"})
		return
	}

	s, err := h.service.Sale(input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, s)
}

// Trade registra un intercambio de stock.
// POST /stock/trade
func (h *StockHandler) Trade(c *gin.Context) {
	var input stock.DecreaseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cuerpo de la petición inválido"})
		return
	}

	s, err := h.service.Trade(input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, s)
}

// Gift registra una donación de stock.
// POST /stock/gift
func (h *StockHandler) Gift(c *gin.Context) {
	var input stock.DecreaseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cuerpo de la petición inválido"})
		return
	}

	s, err := h.service.Gift(input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, s)
}

// Lost registra una pérdida de stock.
// POST /stock/lost
func (h *StockHandler) Lost(c *gin.Context) {
	var input stock.DecreaseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cuerpo de la petición inválido"})
		return
	}

	s, err := h.service.Lost(input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, s)
}

// Damage registra daño a stock.
// POST /stock/damage
func (h *StockHandler) Damage(c *gin.Context) {
	var input stock.DecreaseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cuerpo de la petición inválido"})
		return
	}

	s, err := h.service.Damage(input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, s)
}

// Adjust realiza un ajuste manual de cantidad.
// POST /stock/adjust
func (h *StockHandler) Adjust(c *gin.Context) {
	var input stock.AdjustmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cuerpo de la petición inválido"})
		return
	}

	s, err := h.service.Adjust(input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, s)
}

// Rollback revierte el stock al estado previo a un log específico.
// POST /stock/rollback
func (h *StockHandler) Rollback(c *gin.Context) {
	var input stock.RollbackInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cuerpo de la petición inválido"})
		return
	}

	s, err := h.service.Rollback(input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, s)
}

// GetMyStock obtiene el stock del usuario autenticado.
// GET /stock/me
func (h *StockHandler) GetMyStock(c *gin.Context) {
	userID, _ := c.Get("userID")
	stocks, err := h.service.GetStockByUserID(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stocks)
}

// GetByUserID obtiene todo el stock de un usuario.
// GET /stock/:user_id
func (h *StockHandler) GetByUserID(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id requerido"})
		return
	}

	stocks, err := h.service.GetStockByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stocks)
}

// GetByID obtiene un stock por su ID.
// GET /stock/id/:id
func (h *StockHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	s, err := h.service.GetStockByID(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, s)
}

// GetLogs obtiene el historial de logs de un stock.
// GET /stock/logs/:stock_id
func (h *StockHandler) GetLogs(c *gin.Context) {
	stockIDStr := c.Param("stock_id")
	stockID, err := strconv.ParseUint(stockIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stock_id inválido"})
		return
	}

	logs, err := h.service.GetLogsByStockID(stockID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

func (h *StockHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, stock.ErrStockNotFound),
		errors.Is(err, stock.ErrLogNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, stock.ErrInsufficientStock):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, stock.ErrInvalidQuantity),
		errors.Is(err, stock.ErrInvalidLogType),
		errors.Is(err, stock.ErrRollbackNotAllowed),
		errors.Is(err, stock.ErrStockAlreadyExists):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
