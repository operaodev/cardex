package cards

import (
	"time"
)

type (
	TCG    string
	Rarity string
	Print  string
)

const (
	MTG TCG = "Magic"
	YGO TCG = "Yu-Gi-Oh!"
	PKM TCG = "Pokemon"
)

// Card Code: "RAO5-SP001"
// Wanted: 10234
// TCG: "ygo"
// ExternalID: "ygo-100320"
// EnglishName: "Blue-Eyes White Dragon"
// Name: "Dragón blanco de ojos azules"
// Description: "No puedes invocar esta carta de forma especial desde tu mano, excepto con su propio efecto."
// Lang: "sp"
// Type: "Monster"
// Subtypes: ["Normal"]
// Archetype: "Blue-Eyes"
// Sources: ["yugipedia:url", "ygoprodeck:url"]
// Images: []
// ReferenceImage: ""
// SetName: "Tempestad furiosa"
// SetEnglishName: "Raging Tempest"
// SetCode: "RAO5"
// Rarity: "common"
// Print: "reprint" or "new artwork"
// La identidad única de una carta física es: ExternalID + Code + Lang + Rarity + SetEnglishName.
type Card struct {
	ID         uint64 `json:"id" gorm:"primaryKey;autoIncrement"`
	ExternalID string `json:"external_id" gorm:"not null;uniqueIndex:idx_card_identity,priority:1"`
	Code       string `json:"code,omitempty" gorm:"uniqueIndex:idx_card_identity,priority:2"`
	Wanted     uint   `json:"wanted" gorm:"default:0;not null;index"`
	TCG        TCG    `json:"tcg" gorm:"size:20;not null;index:idx_tcg_lang,priority:1"`

	Name        string      `json:"name" gorm:"not null;index"`
	EnglishName string      `json:"english_name" gorm:"not null;index"`
	Description string      `json:"description" gorm:"not null"`
	Lang        LangCode    `json:"lang" gorm:"size:2;not null;index:idx_tcg_lang,priority:2;uniqueIndex:idx_card_identity,priority:3"`
	Type        string      `json:"type" gorm:"not null;index"`
	Subtypes    []string    `json:"subtypes,omitempty" gorm:"type:jsonb;serializer:json;default:'[]';index:,type:gin"`
	Archetype   string      `json:"archetype,omitempty" gorm:"index"`
	Sources     []string    `json:"sources,omitempty" gorm:"type:jsonb;serializer:json;default:'[]'"`
	CardImages  []CardImage `json:"images,omitempty" gorm:"type:jsonb;serializer:json;default:'[]'"`
	PrintImage  string      `json:"print_image,omitempty" gorm:"index"`

	SetName        string `json:"set_name" gorm:"not null;index"`
	SetEnglishName string `json:"set_english_name" gorm:"not null;uniqueIndex:idx_card_identity,priority:4"`
	SetCode        string `json:"set_code,omitempty" gorm:"index"`
	Rarity         Rarity `json:"rarity,omitempty" gorm:"size:60;uniqueIndex:idx_card_identity,priority:3"`
	Print          Print  `json:"print,omitempty" gorm:"size:30;index"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CardImage struct {
	URL        string `json:"image_url"`
	URLSmall   string `json:"image_url_small"`
	URLCropped string `json:"image_url_cropped"`
}
