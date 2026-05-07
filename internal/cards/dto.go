package cards

type RecomendationCardDTO struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	TCG   TCG    `json:"tcg"`
	Image string `json:"image"`
}

type RecomendationPrintedCardDTO struct {
	ID             uint64   `json:"id"`
	TCG            TCG      `json:"tcg"`
	Name           string   `json:"name"`
	EnglishName    string   `json:"englishName"`
	Code           string   `json:"code"`
	Rarity         Rarity   `json:"rarity"`
	RarityCode     string   `json:"rarityCode"`
	SetName        string   `json:"setName"`
	Image          string   `json:"image"`
	Print          Print    `json:"print"`
}