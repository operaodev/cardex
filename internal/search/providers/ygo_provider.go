package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
	"github.com/operaodev/cardex/internal/cards"
	"github.com/operaodev/cardex/internal/search"
)

// yugipediaLangMap maps the language names used in Yugipedia's wikitable
// to our internal LangCode constants.
var yugipediaLangMap = map[string]cards.LangCode{
	"English":            cards.EN,
	"French":             cards.FR,
	"German":             cards.DE,
	"Italian":            cards.IT,
	"Portuguese":         cards.PT,
	"Spanish":            cards.SP,
	"Japanese":           cards.JP,
	"Korean":             cards.KR,
	"Simplified Chinese": cards.SC,
}

type YGOProCard struct {
	ID        int               `json:"id"`
	Types     string            `json:"humanReadableCardType"`
	Archetype string            `json:"archetype"`
	Images    []cards.CardImage `json:"card_images"`
}

type YGOProvider struct {
	httpClient       *http.Client
	ygoproBaseURL    string
	yugipediaBaseURL string
}

func NewYGOProvider() *YGOProvider {
	return &YGOProvider{
		httpClient:       &http.Client{},
		ygoproBaseURL:    "https://db.ygoprodeck.com/api/v7",
		yugipediaBaseURL: "https://yugipedia.com",
	}
}

func (p *YGOProvider) FetchCardByID(id string) (search.TCGResult, error) {
	reqURL := fmt.Sprintf("%s/cardinfo.php?id=%s", p.ygoproBaseURL, id)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return search.TCGResult{}, err
	}
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return search.TCGResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return search.TCGResult{}, fmt.Errorf("código de estado inesperado: %d", resp.StatusCode)
	}

	var responseYGOPRO struct {
		Data []YGOProCard `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&responseYGOPRO); err != nil {
		return search.TCGResult{}, err
	}

	ygoproCard := responseYGOPRO.Data[0]
	sharedID := cards.GenerateSharedID(cards.TCGTypeYugioh, strconv.Itoa(ygoproCard.ID))

	reqURLYugipedia := fmt.Sprintf("%s/wiki/%s", p.yugipediaBaseURL, id)

	localizedCards, err := scrapeYugipediaCard(reqURLYugipedia)
	if err != nil {
		return search.TCGResult{}, fmt.Errorf("error scraping Yugipedia: %w", err)
	}

	// Build the shared image slice from YGOPro data.
	sharedImages := ygoproCard.Images

	// Enrich each scraped card with the shared fields from YGOPro.
	for lang, card := range localizedCards {
		card.SharedID = sharedID
		card.ID = cards.GenerateCardID(cards.TCGTypeYugioh, lang, strconv.Itoa(ygoproCard.ID))
		card.TCG = cards.TCGTypeYugioh

		splitedTypes := strings.Split(ygoproCard.Types, " ")
		card.Type = splitedTypes[len(splitedTypes)-1] // Usually "Monster", "Spell Card", or "Trap Card"
		if len(splitedTypes) > 1 {
			card.Subtypes = splitedTypes[:len(splitedTypes)-1]
		} else {
			card.Subtypes = []string{}
		}

		card.Archetype = ygoproCard.Archetype
		card.Images = sharedImages
		card.Source = reqURLYugipedia
		localizedCards[lang] = card
	}

	result := search.TCGResult{
		Cards: map[string]map[cards.LangCode]cards.Card{
			sharedID: localizedCards,
		},
	}

	return result, nil
}

func (p *YGOProvider) FetchCards(query string) (search.TCGResult, error) {
	return search.TCGResult{}, nil
}

func scrapeYugipediaCard(url string) (map[cards.LangCode]cards.Card, error) {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	var scrapeErr error
	c.OnError(func(r *colly.Response, err error) {
		scrapeErr = fmt.Errorf("error scraping %s: %w (Status: %d)",
			r.Request.URL, err, r.StatusCode)
	})

	result := make(map[cards.LangCode]cards.Card)

	// Track the current language across rowspan rows.
	var currentLang cards.LangCode
	var currentCardText string

	c.OnHTML("table.wikitable:not(.sortable) tbody tr", func(row *colly.HTMLElement) {
		// The language header cell is a <th scope="row">.
		langCell := strings.TrimSpace(row.ChildText("th[scope=row]"))

		// If this row has a language header, update our tracking variables.
		if langCell != "" {
			code, ok := yugipediaLangMap[langCell]
			if !ok {
				// Unknown language — skip the row.
				currentLang = ""
				return
			}
			currentLang = code

			// The card text may span multiple rows (rowspan); grab it here
			// if present (it won't appear in the second row for JP/KR).
			currentCardText = strings.TrimSpace(row.ChildText("td:last-child"))
		}

		// Skip rows without a recognised language.
		if currentLang == "" {
			return
		}

		// The name is always in the first <td> of the row.
		name := strings.TrimSpace(row.ChildText("td:first-of-type"))
		if name == "" {
			// Second row of a rowspan (romanisation) — nothing new to store.
			return
		}

		langName, _ := cards.GetLangName(currentLang)

		result[currentLang] = cards.Card{
			Lang:        currentLang,
			Language:    langName,
			Name:        name,
			Description: currentCardText,
		}
	})

	if err := c.Visit(url); err != nil {
		return nil, fmt.Errorf("error scraping %s: %w", url, err)
	}
	if scrapeErr != nil {
		return nil, scrapeErr
	}

	return result, nil
}
