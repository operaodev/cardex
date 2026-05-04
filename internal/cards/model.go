package cards

import (
	"fmt"
	"time"
)

type TCGType string
type CardType string

const (
	TCGTypeMagic   TCGType = "MTG"
	TCGTypeYugioh  TCGType = "YGO"
	TCGTypePokemon TCGType = "PKM"
)

// CardInfo almacena los datos invariantes de una carta:
// aquellos que no cambian sin importar el idioma (tipo, subtypes, tags, imágenes, etc.).
// Una misma CardInfo puede tener múltiples Cards asociadas, una por cada idioma disponible.
type CardInfo struct {
	ID        string      `json:"id"         gorm:"primaryKey;size:50"`
	TCG       TCGType     `json:"tcg"        gorm:"size:20;not null;index"`
	Type      CardType    `json:"type"       gorm:"not null;index"`
	Subtypes  []string    `json:"subtypes" gorm:"type:jsonb;serializer:json;default:'[]'"`
	Archetype string      `json:"archetype"  gorm:"index"`
	Source    string      `json:"source"`
	Images    []CardImage `json:"images"     gorm:"type:jsonb;serializer:json;default:'[]'"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// HasMany: una CardInfo tiene muchos Cards (uno por idioma)
	Cards []Card `json:"cards,omitempty" gorm:"foreignKey:CardInfoID"`
}

// Card almacena el contenido localizado de una carta para un idioma específico.
// Cada fila es una traducción: misma carta base, distinto idioma.
// Se relaciona con CardInfo a través de CardInfoID (clave foránea).
type Card struct {
	ID          string      `json:"id"           gorm:"primaryKey;size:50"`
	SharedID    string      `json:"shared_id"    gorm:"not null;index"`
	Lang        LangCode    `json:"lang"         gorm:"size:10;not null;index"`
	Name        string      `json:"name"         gorm:"not null;index"`
	Description string      `json:"description"`
	TCG         TCGType     `json:"tcg"        gorm:"size:20;not null;index"`
	Type        CardType    `json:"type"       gorm:"not null;index"`
	Subtypes    []string    `json:"subtypes"   gorm:"type:jsonb;serializer:json;default:'[]'"`
	Archetype   string      `json:"archetype"  gorm:"index"`
	Source      string      `json:"source"`
	Images      []CardImage `json:"images"     gorm:"type:jsonb;serializer:json;default:'[]'"`
	
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// CardImage almacena las URLs de los distintos tamaños de imagen de una carta.
type CardImage struct {
	URL        string `json:"image_url"`
	URLSmall   string `json:"image_url_small"`
	URLCropped string `json:"image_url_cropped"`
}

// GenerateCardId("YGO", "EN", "12345") // "YGO-EN-12345"
func GenerateCardId(tcg TCGType, lang LangCode, code string) string {
	return fmt.Sprintf("%s-%s-%s", tcg, lang, code)
}

// GenerateCardInfoId("YGO", "12345") // "YGO-12345"
func GenerateCardInfoId(tcg TCGType, code string) string {
	return fmt.Sprintf("%s-%s", tcg, code)
}
