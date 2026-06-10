package products

import "fmt"

// Service define el contrato de operaciones de negocio para products.
type Service interface {
	Upsert(products []Product) (int, error)
	GetByID(id uint64) (*Product, error)
	GetRandomNames(count int) ([]string, error)
	GetSuggestions(input SuggestionInput) ([]SuggestionDTO, error)
	GetSuggestionsByUser(userID string, input SuggestionInput) ([]SuggestionDTO, error)
	GetRelatedCards(input RelatedCardsInput) (*RelatedCardsResponse, error)
	GetCardsBySet(input GetCardsBySetInput) ([]Product, error)
}

// service implementa la interfaz Service e inyecta el repositorio.
type service struct {
	repo Repository
}

// NewService crea una nueva instancia del servicio de products.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// Upsert inserta o actualiza products en lote.
func (s *service) Upsert(products []Product) (int, error) {
	if len(products) == 0 {
		return 0, fmt.Errorf("la lista de products no puede estar vacía")
	}
	return s.repo.Upsert(products)
}

// GetByID obtiene un product por su ID e incrementa el contador wanted.
func (s *service) GetByID(id uint64) (*Product, error) {
	if id == 0 {
		return nil, fmt.Errorf("el ID no puede estar vacío")
	}
	return s.repo.GetByID(id)
}

// GetRandomNames obtiene nombres de products aleatorios.
func (s *service) GetRandomNames(count int) ([]string, error) {
	if count <= 0 {
		return nil, fmt.Errorf("la cantidad debe ser mayor a cero")
	}
	if count > 50 {
		count = 50
	}
	return s.repo.GetRandomNames(count)
}

// GetSuggestions obtiene sugerencias de products por búsqueda parcial.
func (s *service) GetSuggestions(input SuggestionInput) ([]SuggestionDTO, error) {
	if input.Input == "" {
		return nil, fmt.Errorf("el término de búsqueda no puede estar vacío")
	}
	if len(input.Input) < 2 {
		return nil, fmt.Errorf("el término de búsqueda debe tener al menos 2 caracteres")
	}
	return s.repo.GetSuggestions(input)
}

func (s *service) GetSuggestionsByUser(userID string, input SuggestionInput) ([]SuggestionDTO, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID requerido")
	}
	if input.Input == "" {
		return nil, fmt.Errorf("el término de búsqueda no puede estar vacío")
	}
	if len(input.Input) < 2 {
		return nil, fmt.Errorf("el término de búsqueda debe tener al menos 2 caracteres")
	}
	return s.repo.GetSuggestionsByUser(userID, input)
}

func (s *service) GetRelatedCards(input RelatedCardsInput) (*RelatedCardsResponse, error) {
	if input.ID == 0 {
		return nil, fmt.Errorf("el ID no puede estar vacío")
	}
	return s.repo.GetRelatedCards(input)
}

func (s *service) GetCardsBySet(input GetCardsBySetInput) ([]Product, error) {
	if input.SetExternalID == "" {
		return nil, fmt.Errorf("set_external_id no puede estar vacío")
	}
	if input.Lang == "" {
		return nil, fmt.Errorf("lang no puede estar vacío")
	}
	return s.repo.GetCardsBySet(input.SetExternalID, LangCode(input.Lang))
}
