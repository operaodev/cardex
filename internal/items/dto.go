package items

type SuggestionDTO struct {
	ID            uint64   `json:"id"`
	ExternalID    string   `json:"external_id"`
	SetExternalID string   `json:"set_external_id"`
	Type          ItemType `json:"type"`
	TCG           TCG      `json:"tcg"`
	Wanted        uint     `json:"wanted"`
	Name          string   `json:"name"`
	Code          string   `json:"code,omitempty"`
	Rarity        string   `json:"rarity,omitempty"`
	RarityCode    string   `json:"rarity_code,omitempty"`
	SetName       string   `json:"set_name,omitempty"`
	SetCode       string   `json:"set_code,omitempty"`
	Lang          string   `json:"lang,omitempty"`
	Language      LangCode `json:"language,omitempty"`
	Image         string   `json:"image,omitempty"`
	Edition       string   `json:"edition,omitempty"`
}
