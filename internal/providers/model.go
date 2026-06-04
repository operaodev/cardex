package providers

import (
	"github.com/operaodev/cardex/internal/items"
)

type Provider interface {
	FetchSets(cards []items.Item) ([]items.Item, error)
	FetchCards() ([]items.Item, error)
	FetchCardsByName(name string) ([]items.Item, error)
}
