package marketplace

import "fmt"

type Service interface {
	GetPrices(id uint64) (MarketAnalysis, error)
	GetOffers(input OffersInput) (OffersPage, error)
	GetCards(input FilterInput) (ProductResumePage, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetPrices(id uint64) (MarketAnalysis, error) {
	if id == 0 {
		return MarketAnalysis{}, fmt.Errorf("el ID no puede estar vacío")
	}
	return s.repo.GetPrices(id)
}

func (s *service) GetOffers(input OffersInput) (OffersPage, error) {
	if input.ProductID == 0 {
		return OffersPage{}, fmt.Errorf("el ID no puede estar vacío")
	}
	return s.repo.GetOffers(input)
}

func (s *service) GetCards(input FilterInput) (ProductResumePage, error) {
	// if input.Input == "" {
	// 	return ProductResumePage{}, fmt.Errorf("el término de búsqueda no puede estar vacío")
	// }
	// if len(input.Input) < 2 {
	// 	return ProductResumePage{}, fmt.Errorf("el término de búsqueda debe tener al menos 2 caracteres")
	// }
	return s.repo.GetCards(input)
}
