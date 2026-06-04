package items

import "fmt"

// Service define el contrato de operaciones de negocio para ítems.
type Service interface {
	Upsert(items []Item) (int, error)
	GetByID(id uint64) (*Item, error)
	GetRandomNames(count int) ([]string, error)
	GetSuggestions(input SuggestionInput) ([]SuggestionDTO, error)
}

// service implementa la interfaz Service e inyecta el repositorio.
type service struct {
	repo Repository
}

// NewService crea una nueva instancia del servicio de ítems.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// Upsert inserta o actualiza ítems en lote.
func (s *service) Upsert(items []Item) (int, error) {
	if len(items) == 0 {
		return 0, fmt.Errorf("la lista de ítems no puede estar vacía")
	}
	return s.repo.Upsert(items)
}

// GetByID obtiene un ítem por su ID e incrementa el contador wanted.
func (s *service) GetByID(id uint64) (*Item, error) {
	if id == 0 {
		return nil, fmt.Errorf("el ID no puede estar vacío")
	}
	return s.repo.GetByID(id)
}

// GetRandomNames obtiene nombres de ítems aleatorios.
func (s *service) GetRandomNames(count int) ([]string, error) {
	if count <= 0 {
		return nil, fmt.Errorf("la cantidad debe ser mayor a cero")
	}
	if count > 50 {
		count = 50
	}
	return s.repo.GetRandomNames(count)
}

// GetSuggestions obtiene sugerencias de ítems por búsqueda parcial.
func (s *service) GetSuggestions(input SuggestionInput) ([]SuggestionDTO, error) {
	if input.Input == "" {
		return nil, fmt.Errorf("el término de búsqueda no puede estar vacío")
	}
	if len(input.Input) < 2 {
		return nil, fmt.Errorf("el término de búsqueda debe tener al menos 2 caracteres")
	}
	return s.repo.GetSuggestions(input)
}
