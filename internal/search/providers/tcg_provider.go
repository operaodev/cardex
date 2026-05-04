package search

type TCGProvider[T any, K any] interface {
	FetchCards(key K) ([]T, error)
}
