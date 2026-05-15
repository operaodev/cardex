package search

import (
	"fmt"
)

type Service struct {
	ygo TCGProvider
}

func NewService(ygo TCGProvider) *Service {
	return &Service{
		ygo: ygo,
	}
}

func (s *Service) SearchAll(providerName string) ([]ResultCard, error) {
	switch providerName {
	case "ygo":
		result, err := s.ygo.FetchAllCards()
		if err != nil {
			return nil, fmt.Errorf("error obteniendo cartas: %w", err)
		}
		fmt.Printf("Encontradas %d cartas en Yugipedia\n", len(result))
		return result, nil
	default:
		return nil, fmt.Errorf("proveedor no soportado: %s", providerName)
	}
}

// SearchByID fetches a single card by its provider ID and returns all
// localised versions as a flat slice.
func (s *Service) SearchByID(providerName, id string) (ResultCard, error) {
	switch providerName {
	case "ygo":
		result, err := s.ygo.FetchCardByID(id)
		if err != nil {
			return ResultCard{}, fmt.Errorf("error obteniendo carta: %w", err)
		}
		return result, nil
	default:
		return ResultCard{}, fmt.Errorf("proveedor no soportado: %s", providerName)
	}
}

func (s *Service) SearchByNames(providerName, name string) ([]ResultCard, error) {
	switch providerName {
	case "ygo":
		results, err := s.ygo.FetchCardsByName(name)
		if err != nil {
			return nil, fmt.Errorf("error obteniendo cartas: %w", err)
		}
		return results, nil
	default:
		return nil, fmt.Errorf("proveedor no soportado: %s", providerName)
	}
}
