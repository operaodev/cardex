package custompacks

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	Upsert(userID string, productID uint64, delta int) (*Wishlist, error)
	Delete(userID string, productID uint64) error
	GetByUserID(userID string) ([]Wishlist, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Upsert incrementa/decrementa la cantidad de un item en la wishlist.
// Si la cantidad resultante es <= 0, lo elimina.
func (r *repository) Upsert(userID string, productID uint64, delta int) (*Wishlist, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID no puede estar vacío")
	}
	if productID == 0 {
		return nil, fmt.Errorf("productID no puede ser 0")
	}

	// Intentamos insertar o actualizar
	wish := Wishlist{
		UserID:    userID,
		ProductID: productID,
		Quantity:  delta,
	}

	result := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_id"},
			{Name: "product_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"quantity":   gorm.Expr("wishlists.quantity + ?", delta),
			"updated_at": gorm.Expr("NOW()"),
		}),
	}).Create(&wish)

	if result.Error != nil {
		return nil, result.Error
	}

	// Recargamos para obtener el valor real post-upsert
	if err := r.db.First(&wish, wish.ID).Error; err != nil {
		return nil, err
	}

	// Si la cantidad bajó a 0 o menos, eliminamos
	if wish.Quantity <= 0 {
		if err := r.db.Delete(&wish).Error; err != nil {
			return nil, err
		}
		return nil, nil
	}

	return &wish, nil
}

// Delete elimina un item de la wishlist.
func (r *repository) Delete(userID string, productID uint64) error {
	result := r.db.
		Where("user_id = ?", userID).
		Where("product_id = ?", productID).
		Delete(&Wishlist{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("item no encontrado")
	}
	return nil
}

// GetByUserID obtiene todos los items de la wishlist de un usuario.
func (r *repository) GetByUserID(userID string) ([]Wishlist, error) {
	var items []Wishlist
	result := r.db.
		Preload("Product").
		Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&items)

	if result.Error != nil {
		return nil, result.Error
	}
	return items, nil
}
