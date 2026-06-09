package custompacks

import (
	"time"

	"github.com/operaodev/cardex/internal/products"
	"github.com/operaodev/cardex/internal/users"
)

type Wishlist struct {
	ID        uint64 `json:"id"         gorm:"primaryKey;autoIncrement"`
	UserID    string `json:"user_id"    gorm:"not null;type:uuid;uniqueIndex:idx_wishlist_user_product,priority:1"`
	ProductID uint64 `json:"product_id" gorm:"not null;uniqueIndex:idx_wishlist_user_product,priority:2"`
	Quantity  int    `json:"quantity"   gorm:"default:1;not null"`

	User    users.User       `json:"-"       gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	Product products.Product `json:"product"  gorm:"foreignKey:ProductID;references:ID;constraint:OnDelete:RESTRICT"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
