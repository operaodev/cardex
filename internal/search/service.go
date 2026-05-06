package search

import (
	"fmt"

	"github.com/operaodev/cardex/internal/cards"
)

type Service struct {
	ygo TCGProvider
}

func NewService(ygo TCGProvider) *Service {
	return &Service{
		ygo: ygo,
	}
}

// SearchByProvider fetches cards from the given provider and flattens the
// multilingual TCGResult into a slice of cards.Card for the HTTP response.
func (s *Service) SearchByProvider(providerName, query string) ([]cards.Card, error) {
	switch providerName {
	case "ygo":
		result, err := s.ygo.FetchCards(query)
		if err != nil {
			return nil, fmt.Errorf("error obteniendo cartas: %w", err)
		}
		return flattenTCGResult(result), nil
	default:
		return nil, fmt.Errorf("proveedor no soportado: %s", providerName)
	}
}

// SearchByID fetches a single card by its provider ID and returns all
// localised versions as a flat slice.
func (s *Service) SearchByID(providerName, id string) ([]cards.Card, error) {
	switch providerName {
	case "ygo":
		result, err := s.ygo.FetchCardByID(id)
		if err != nil {
			return nil, fmt.Errorf("error obteniendo carta: %w", err)
		}
		return flattenTCGResult(result), nil
	default:
		return nil, fmt.Errorf("proveedor no soportado: %s", providerName)
	}
}

// flattenTCGResult converts the nested map[sharedID]map[LangCode]Card into
// a flat []cards.Card for easy JSON serialisation.
func flattenTCGResult(result TCGResult) []cards.Card {
	var out []cards.Card
	for _, byLang := range result.Cards {
		for _, card := range byLang {
			out = append(out, card)
		}
	}
	return out
}
