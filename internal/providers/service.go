package providers

import (
	"fmt"

	"github.com/operaodev/cardex/internal/items"
)

type Service struct {
	ygo *YGOProvider
}

func NewService(ygo *YGOProvider) *Service {
	return &Service{
		ygo: ygo,
	}
}

// FetchCardsByName fetches cards matching name from the given provider.
// If providerName is empty, all providers are queried and results are merged.
func (s *Service) FetchCardsByName(provider items.TCG, name string) ([]items.Item, error) {
	if provider == "" {
		return s.searchAllProvidersByName(name)
	}

	switch provider {
	case items.YGO:
		results, err := s.ygo.FetchCardsByName(name)
		if err != nil {
			return nil, fmt.Errorf("ygo: error fetching cards by name: %w", err)
		}
		return results, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// FetchCardsByName fetches every available card from the given provider.
// If providerName is empty, all providers are queried and results are merged.
func (s *Service) FetchCards(provider items.TCG) ([]items.Item, error) {
	if provider == "" {
		return s.searchAllProviders()
	}

	switch provider {
	case items.YGO:
		results, err := s.ygo.FetchCards()
		if err != nil {
			return nil, fmt.Errorf("ygo: error fetching all cards: %w", err)
		}
		fmt.Printf("[ygo] found %d cards\n", len(results))
		return results, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// searchAllProviders runs SearchAll across every provider concurrently.
func (s *Service) searchAllProviders() ([]items.Item, error) {
	var all []items.Item
	var errs []error

	// Add each provider call here when new providers are registered.
	for _, call := range s.allFetchers() {
		got, err := call()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		all = append(all, got...)
	}

	if len(errs) > 0 {
		return all, fmt.Errorf("provider errors: %v", errs)
	}
	return all, nil
}

// searchAllProvidersByName runs SearchByName across every provider concurrently.
func (s *Service) searchAllProvidersByName(name string) ([]items.Item, error) {
	var all []items.Item
	var errs []error

	for _, call := range s.allNameFetchers(name) {
		got, err := call()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		all = append(all, got...)
	}

	if len(errs) > 0 {
		return all, fmt.Errorf("provider errors: %v", errs)
	}
	return all, nil
}

// allFetchers returns a FetchCards call per active provider.
// Add new providers here as they are implemented.
func (s *Service) allFetchers() []func() ([]items.Item, error) {
	return []func() ([]items.Item, error){
		s.ygo.FetchCards,
		// s.mtg.FetchCards,
	}
}

// allNameFetchers returns a FetchCardsByName call per active provider.
func (s *Service) allNameFetchers(name string) []func() ([]items.Item, error) {
	return []func() ([]items.Item, error){
		func() ([]items.Item, error) { return s.ygo.FetchCardsByName(name) },
		// func() ([]items.Item, error) { return s.mtg.FetchCardsByName(name) },
	}
}
