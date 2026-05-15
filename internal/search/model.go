package search

import "github.com/operaodev/cardex/internal/cards"

type ResultPrintedCard struct {
	Code    string         `json:"code,omitempty"`
	Rarity  cards.Rarity   `json:"rarity,omitempty"`
	SetName string         `json:"setName,omitempty"`
	Lang    cards.LangCode `json:"lang,omitempty"`
	TCG     cards.TCG      `json:"tcg,omitempty"`
}

type ResultCard struct {
	ExternalID   string                    `json:"externalId,omitempty"`
	TCG          cards.TCG                 `json:"tcg,omitempty"`
	Names        map[cards.LangCode]string `json:"names,omitempty"`
	Descriptions map[cards.LangCode]string `json:"descriptions,omitempty"`
	Type         string                    `json:"type,omitempty"`
	Subtypes     []string                  `json:"subtypes,omitempty"`
	Archetype    string                    `json:"archetype,omitempty"`
	Sources      []string                  `json:"sources,omitempty"`
	Images       []cards.CardImage         `json:"images,omitempty"`
	PrintedCards []ResultPrintedCard `json:"printedCards,omitempty"`
}

type TCGProvider interface {
	FetchAllCards() ([]ResultCard, error)
	FetchCardByID(id string) (ResultCard, error)
	FetchCardsByName(name string) ([]ResultCard, error)
}
