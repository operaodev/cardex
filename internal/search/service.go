package search

import (
	"fmt"

	"github.com/operaodev/cardex/internal/cards"
)

type Service struct {
	ygo TCGProvider[YGOPROCard, string]
}

func NewService(ygo TCGProvider[YGOPROCard, string]) *Service {
	return &Service{
		ygo: ygo,
	}
}

func (s *Service) SearchByProvider(providerName, query string) ([]cards.Card, error) {
	switch providerName {
	case "ygopro":
		return executeSearch(s.ygo, query)
	default:
		return nil, fmt.Errorf("proveedor no soportado: %s", providerName)
	}
}

// Función auxiliar genérica que maneja la lógica repetitiva para cualquier TCGProvider
func executeSearch[T any](p searchproviders.TCGProvider[T, string], query string) ([]cards.Card, error) {
	rawCards, err := p.FetchCards(query)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo cartas: %w", err)
	}

	var results []cards.Card
	for _, raw := range rawCards {
		card, err := p.ConvertToCard(raw)
		if err != nil {
			continue
		}
		results = append(results, card)
	}
	return results, nil
}
