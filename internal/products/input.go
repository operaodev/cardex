package products

type SuggestionInput struct {
	TCG   TCG      `json:"tcg"`
	Lang  LangCode `json:"lang"`
	Input string   `json:"input"`
}

type CatalogInput struct {
	Input string `json:"input"`
	TCG   TCG    `json:"tcg"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}
