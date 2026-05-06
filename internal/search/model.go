package search

import "github.com/operaodev/cardex/internal/cards"

type TCGResult struct {
	Cards map[string]map[cards.LangCode]cards.Card
}

type TCGProvider interface {
	FetchCardByID(id string) (TCGResult, error)
	FetchCards(query string) (TCGResult, error)
}
