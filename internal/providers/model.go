package providers

import (
	"github.com/operaodev/cardex/internal/products"
)

type Provider interface {
	FetchItems() ([]products.Product, error)
	FetchItemsByName(name string) ([]products.Product, error)
}
