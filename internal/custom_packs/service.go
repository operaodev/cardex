package custompacks

import "fmt"

// Service define el contrato de operaciones de negocio para wishlist.
type Service interface {
	Upsert(userID string, productID uint64, delta int) (*Wishlist, error)
	Delete(userID string, productID uint64) error
	GetByUserID(userID string) ([]Wishlist, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Upsert(userID string, productID uint64, delta int) (*Wishlist, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID requerido")
	}
	if productID == 0 {
		return nil, fmt.Errorf("productID requerido")
	}
	if delta == 0 {
		return nil, fmt.Errorf("delta no puede ser 0")
	}
	return s.repo.Upsert(userID, productID, delta)
}

func (s *service) Delete(userID string, productID uint64) error {
	if userID == "" {
		return fmt.Errorf("userID requerido")
	}
	if productID == 0 {
		return fmt.Errorf("productID requerido")
	}
	return s.repo.Delete(userID, productID)
}

func (s *service) GetByUserID(userID string) ([]Wishlist, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID requerido")
	}
	return s.repo.GetByUserID(userID)
}
