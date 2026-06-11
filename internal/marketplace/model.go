package marketplace

import (
	"github.com/operaodev/cardex/internal/stock"
	"github.com/operaodev/cardex/internal/users"
)

type ProductResume struct {
	ID           uint64  `json:"id"`
	Name         string  `json:"name"`
	Code         string  `json:"code,omitempty"`
	SetName      string  `json:"set_name,omitempty"`
	Rarity       string  `json:"rarity,omitempty"`
	RarityCode   string  `json:"rarity_code,omitempty"`
	Image        string  `json:"image,omitempty"`
	GlobalStock  int     `json:"global_stock,omitempty"`
	AveragePrice float64 `json:"average_price,omitempty"`
}

type FilterInput struct {
	Input       string   `json:"input"`
	ProductType string   `json:"product_type"`
	TCGs        []string `json:"tcgs"`
	Langs       []string `json:"langs"`
	Page        int      `json:"page"`
	Limit       int      `json:"limit"`
}

type ProductResumePage struct {
	Items      []ProductResume `json:"items"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"total_pages"`
}

type MarketAnalysis struct {
	ProductId    uint64  `json:"product_id"`
	LowPrice     float64 `json:"low_price"`
	HighPrice    float64 `json:"high_price"`
	AveragePrice float64 `json:"average_price"`
	MarketStocks uint    `json:"market_stock"`
}

type Offer struct {
	User          users.User      `json:"user"`
	StockID       uint64          `json:"stock_id"`
	Condition     stock.Condition `json:"condition"`
	IsForTrade    bool            `json:"is_for_trade"`
	Price         float64         `json:"price"`
	DiscountPrice float64         `json:"discount_price"`
	Discount      float64         `json:"discount"`
	Quantity      uint            `json:"quantity"`
}

type OffersInput struct {
	ProductID uint64
	ForSale   *bool
	ForTrade  *bool
	HasStock  *bool
	SortDesc  bool
	Page      int
	Limit     int
}

type OffersPage struct {
	Items      []Offer `json:"items"`
	Total      int64   `json:"total"`
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
	TotalPages int     `json:"total_pages"`
}
