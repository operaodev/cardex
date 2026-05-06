package cards

import (
	"time"
)

type Rarity string

type PrintCard struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement"`
	UniqueID    string `gorm:"not null;index"`
	SharedID    string `gorm:"not null;index"`
	DisplayCode string `gorm:"not null;index"`
	SetCode     string `gorm:"not null;index"`
	SetNumber   string `gorm:"not null;index"`
	Lang        LangCode `gorm:"size:2;not null;index"`
	Language    LangName `gorm:"size:50;not null;index"`

	TCG TCGType `gorm:"size:20;not null;index"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type TCGType string

const (
	TCGTypeMagic   TCGType = "mtg"
	TCGTypeYugioh  TCGType = "ygo"
	TCGTypePokemon TCGType = "pkm"
)

// Card almacena el contenido localizado de una carta para un idioma específico.
// Cada fila es una traducción: misma carta base, distinto idioma.
type Card struct {
	ID          uint64      `json:"id"           gorm:"primaryKey;autoIncrement"`
	UniqueID    string      `json:"unique_id"    gorm:"not null;index"`
	SharedID    string      `json:"shared_id"    gorm:"not null;index"`
	Lang        LangCode    `json:"lang"         gorm:"size:2;not null;index"`
	Language    LangName    `json:"language"     gorm:"size:50;not null;index"`
	Name        string      `json:"name"         gorm:"not null;index"`
	Description string      `json:"description"`
	TCG         TCGType     `json:"tcg"        gorm:"size:20;not null;index"`
	Type        string      `json:"type"       gorm:"not null;index"`
	Subtypes    []string    `json:"subtypes"   gorm:"type:jsonb;serializer:json;default:'[]'"`
	Archetype   string      `json:"archetype"  gorm:"index"`
	Source      string      `json:"source"`
	Images      []CardImage `json:"images"     gorm:"type:jsonb;serializer:json;default:'[]'"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CardImage struct {
	URL        string `json:"image_url"`
	URLSmall   string `json:"image_url_small"`
	URLCropped string `json:"image_url_cropped"`
}
