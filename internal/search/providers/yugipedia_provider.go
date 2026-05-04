package search

import (
	"net/http"

	"github.com/operaodev/cardex/internal/cards"
)

type YugipediaData struct {
	Cards map[cards.LangCode]YugipediaCard
}

type YugipediaCard struct {
	ID          int
	Name        string
	Description string
	Types       string
	Archetype   string
	Source      string
	CardPrints  []CardPrint
}

type CardPrint struct {
	SetCode    string
	Set        string
	Number     string
	Rarity     string
	RarityCode string
}

type YugipediaProvider struct {
	httpClient *http.Client
	baseUrl    string
}

func NewYugipediaProvider() *YugipediaProvider {
	return &YugipediaProvider{
		httpClient: &http.Client{},
		baseUrl:    "https://yugipedia.com/api/v1",
	}
}