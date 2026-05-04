package search

type CardResult struct {
	ID          int
	Name        string
	Description string
	Types       string
	Archetype   string
	Source      string
	Images      []string
}

type TCGProvider[T any, K any] interface {
	FetchCards(key K) ([]T, error)
}
