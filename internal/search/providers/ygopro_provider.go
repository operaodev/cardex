package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type YGOPROProvider struct {
	httpClient *http.Client
	baseUrl    string
}

func NewYGOPROProvider() *YGOPROProvider {
	return &YGOPROProvider{
		httpClient: &http.Client{},
		baseUrl:    "https://db.ygoprodeck.com/api/v7",
	}
}

func (p *YGOPROProvider) FetchCards(name string) ([]YGOPROCard, error) {
	var reqURL string
	if strings.TrimSpace(name) == "" {
		reqURL = fmt.Sprintf("%s/cardinfo.php", p.baseUrl)
	} else {
		encodedName := url.QueryEscape(name)
		reqURL = fmt.Sprintf("%s/cardinfo.php?fname=%s", p.baseUrl, encodedName)
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error al crear request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar request a ygoprodeck: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusBadRequest {
			return []YGOPROCard{}, nil
		}
		return nil, fmt.Errorf("código de estado inesperado: %d", resp.StatusCode)
	}

	var responseData struct {
		Data []YGOPROCard `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("error al decodificar JSON: %w", err)
	}

	return responseData.Data, nil
}

// func (p *YGOPROProvider) ConvertToCard(data YGOPROCard) (cards.Card, error) {
// 	if data.ID == 0 {
// 		return cards.Card{}, fmt.Errorf("carta invalida: id es 0")
// 	}
// 	if data.Name == "" {
// 		return cards.Card{}, fmt.Errorf("carta invalida: nombre es invalido")
// 	}
// 	if data.Types == "" {
// 		return cards.Card{}, fmt.Errorf("carta invalida: tipo es invalido")
// 	}

// 	var card cards.Card
// 	var info cards.CardInfo

// 	infoID := cards.GenerateCardInfoId(cards.TCGTypeYugioh, fmt.Sprintf("%d", data.ID))
// 	info.ID = infoID
// 	info.TCG = cards.TCGTypeYugioh
	
// 	types := strings.Split(data.Types, " ")
// 	info.Type = cards.CardType(types[len(types)-1])
// 	if len(types) > 1 {
// 		info.Subtypes = types[:len(types)-1]
// 	} else {
// 		info.Subtypes = []string{}
// 	}
	
// 	info.Images = data.Images
// 	info.Source = data.Source
// 	info.Archetype = data.Archetype

// 	card.ID = cards.GenerateCardId(cards.TCGTypeYugioh, cards.EN, fmt.Sprintf("%d", data.ID))
// 	card.Info = info
// 	card.CardInfoID = infoID
// 	card.Name = data.Name
// 	card.Description = data.Description
// 	card.Lang = cards.EN

// 	return card, nil
// }
