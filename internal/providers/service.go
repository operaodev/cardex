package providers

import (
	"fmt"

	"github.com/operaodev/cardex/internal/products"
)

type Service struct {
	ygo *YGOProvider
}

func NewService(ygo *YGOProvider) *Service {
	return &Service{
		ygo: ygo,
	}
}

func (s *Service) FetchItemsByName(provider products.TCG, name string) ([]products.Product, error) {
	switch provider {
	case products.YGO:
		return s.ygo.FetchItemsByName(name)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (s *Service) FetchItems(provider products.TCG) ([]products.Product, error) {
	switch provider {
	case products.YGO:
		return s.ygo.FetchItems()
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}
