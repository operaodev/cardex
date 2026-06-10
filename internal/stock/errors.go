package stock

import "errors"

var (
	ErrStockNotFound       = errors.New("Stock no encontrado")
	ErrLogNotFound         = errors.New("Log no encontrado")
	ErrInsufficientStock   = errors.New("Stock insuficiente")
	ErrInvalidQuantity     = errors.New("Cantidad inválida")
	ErrInvalidLogType      = errors.New("Tipo de log inválido para esta operación")
	ErrRollbackNotAllowed  = errors.New("No se puede hacer rollback de este tipo de log")
	ErrStockAlreadyExists  = errors.New("Ya existe un stock con esta combinación de usuario, producto y condición")
	ErrProductNotASet      = errors.New("El producto no es un set")
	ErrProductNotFromSet   = errors.New("El producto no pertenece al set")
)
