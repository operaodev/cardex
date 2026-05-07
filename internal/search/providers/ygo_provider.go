package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/operaodev/cardex/internal/cards"
	"github.com/operaodev/cardex/internal/search"
)

type YGOProCard struct {
	ID            int               `json:"id"`
	Name          string            `json:"name"`
	Types         string            `json:"humanReadableCardType"`
	Archetype     string            `json:"archetype"`
	Images        []cards.CardImage `json:"card_images"`
	YgoprodeckURL string            `json:"ygoprodeck_url"`
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

func (p *YGOProvider) FetchCardByID(id string) (search.ResultCard, error) {
	reqURL := fmt.Sprintf("%s/cardinfo.php?id=%s", p.ygoproBaseURL, id)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return search.ResultCard{}, err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return search.ResultCard{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return search.ResultCard{},
			fmt.Errorf("código de estado inesperado: %d", resp.StatusCode)
	}

	var responseYGOPRO struct {
		Data []YGOProCard `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&responseYGOPRO); err != nil {
		return search.ResultCard{}, err
	}

	ygoproResult := responseYGOPRO.Data[0]

	yugipediaResults, err := p.getTotalCardInfo(ygoproResult.Name)
	if err != nil {
		return search.ResultCard{},
			fmt.Errorf("error obteniendo info de Yugipedia: %w", err)
	}

	// Usar el primer (y único) resultado de Yugipedia como base
	var result search.ResultCard
	if len(yugipediaResults) > 0 {
		result = yugipediaResults[0]
	}

	// Completar con los datos base de YGOPRO que Yugipedia no provee
	result.ExternalID = fmt.Sprintf("%d", ygoproResult.ID)
	types := strings.Split(ygoproResult.Types, " ")
	result.Type = types[len(types)-1]
	result.Subtypes = types[:len(types)-1]
	result.Archetype = ygoproResult.Archetype
	result.Images = ygoproResult.Images
	result.TCG = cards.TCGYugioh
	result.Sources = append(result.Sources, ygoproResult.YgoprodeckURL)
	result.Sources = append(result.Sources, "https://yugipedia.com/wiki/"+url.QueryEscape(ygoproResult.Name))

	// El nombre en inglés viene de YGOPRO (es la fuente principal en EN)
	if result.Names == nil {
		result.Names = make(map[cards.LangCode]string)
	}
	result.Names[cards.EN] = ygoproResult.Name

	return result, nil
}

func (p *YGOProvider) FetchCardsByName(name string) ([]search.ResultCard, error) {
	// YGOPRO acepta búsqueda fuzzy por nombre con el parámetro fname
	reqURL := fmt.Sprintf("%s/cardinfo.php?fname=%s", p.ygoproBaseURL, url.QueryEscape(name))

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("código de estado inesperado de YGOPRODeck: %d", resp.StatusCode)
	}

	var responseYGOPRO struct {
		Data []YGOProCard `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&responseYGOPRO); err != nil {
		return nil, err
	}

	if len(responseYGOPRO.Data) == 0 {
		return []search.ResultCard{}, nil
	}

	// Recolectar todos los nombres para hacer una sola request a Yugipedia
	// Limitar a 20 nombres porque la API de Yugipedia devuelve error 414 si se envían demasiados
	limit := min(len(responseYGOPRO.Data), 20)
	responseYGOPRO.Data = responseYGOPRO.Data[:limit]

	allNames := make([]string, limit)
	for i, c := range responseYGOPRO.Data {
		allNames[i] = c.Name
	}

	yugipediaResults, err := p.getTotalCardInfo(allNames...)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo info de Yugipedia: %w", err)
	}

	// Indexar resultados de Yugipedia por nombre EN para cruzarlos
	yugipediaByName := make(map[string]search.ResultCard, len(yugipediaResults))
	for _, yResult := range yugipediaResults {
		if enName, ok := yResult.Names[cards.EN]; ok {
			yugipediaByName[enName] = yResult
		}
	}

	results := make([]search.ResultCard, 0, len(responseYGOPRO.Data))

	for _, ygoproCard := range responseYGOPRO.Data {
		result, ok := yugipediaByName[ygoproCard.Name]
		if !ok {
			result = search.ResultCard{}
		}

		// Completar con los datos base de YGOPRO que Yugipedia no provee
		result.ExternalID = fmt.Sprintf("%d", ygoproCard.ID)
		types := strings.Split(ygoproCard.Types, " ")
		result.Type = types[len(types)-1]
		result.Subtypes = types[:len(types)-1]
		result.Archetype = ygoproCard.Archetype
		result.Images = ygoproCard.Images
		result.TCG = cards.TCGYugioh
		result.Sources = append(result.Sources, ygoproCard.YgoprodeckURL)
		result.Sources = append(result.Sources, "https://yugipedia.com/wiki/"+url.QueryEscape(ygoproCard.Name))

		if result.Names == nil {
			result.Names = make(map[cards.LangCode]string)
		}
		result.Names[cards.EN] = ygoproCard.Name

		results = append(results, result)
	}

	return results, nil
}

// getTotalCardInfo obtiene el wikitext de una o más cartas desde Yugipedia
// y extrae los nombres y descripciones localizados de cada una.
func (p *YGOProvider) getTotalCardInfo(names ...string) ([]search.ResultCard, error) {
	// Usar prop=revisions con rvprop=content para obtener el wikitext de la carta.
	titles := url.QueryEscape(strings.Join(names, "|"))
	reqUrl := fmt.Sprintf(
		"%s/api.php?action=query&prop=revisions&rvprop=content&format=json&titles=%s",
		p.yugipediaBaseURL,
		titles,
	)

	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://yugipedia.com/")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil,
			fmt.Errorf("código de estado inesperado de Yugipedia: %d", resp.StatusCode)
	}

	// La API de Mediawiki devuelve pages como un objeto (mapa) keyed por page ID,
	// no como un array, por eso usamos map[string].
	// Sin rvslots, el contenido viene directo en revisions[]["*"].
	var response struct {
		Query struct {
			Pages map[string]struct {
				Title     string `json:"title"`
				Revisions []struct {
					Content string `json:"*"`
				} `json:"revisions"`
			} `json:"pages"`
		} `json:"query"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	totalCards := make([]search.ResultCard, 0, len(response.Query.Pages))

	for _, page := range response.Query.Pages {
		if len(page.Revisions) == 0 {
			continue
		}

		content := strings.TrimSpace(page.Revisions[0].Content)
		card := parseWikitext(content)

		if card.Names == nil {
			card.Names = make(map[cards.LangCode]string)
		}
		card.Names[cards.EN] = page.Title

		totalCards = append(totalCards, card)
	}

	return totalCards, nil
}

// extractLocalizedValues busca claves como:
// | fr_name = ...
// | es_text = ...
func extractLocalizedValues(content string, keys []cards.LangCode) map[cards.LangCode]string {
	result := make(map[cards.LangCode]string)

	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" || !strings.Contains(line, "=") {
			continue
		}

		for _, key := range keys {
			keyStr := string(key)
			prefix := "| " + keyStr

			if strings.HasPrefix(line, prefix) {
				parts := strings.SplitN(line, "=", 2)

				if len(parts) != 2 {
					continue
				}

				value := strings.TrimSpace(parts[1])

				if keyStr == "text" {
					result[cards.EN] = value
				} else {
					result[key] = value
				}
			}
		}
	}

	return result
}

// func scrapeYugipediaCard(url string) (map[cards.LangCode]cards.Card, error) {
// 	c := colly.NewCollector(
// 		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
// 	)

// 	var scrapeErr error
// 	c.OnError(func(r *colly.Response, err error) {
// 		scrapeErr = fmt.Errorf("error scraping %s: %w (Status: %d)",
// 			r.Request.URL, err, r.StatusCode)
// 	})

// 	result := make(map[cards.LangCode]cards.Card)

// 	// Track the current language across rowspan rows.
// 	var currentLang cards.LangCode
// 	var currentCardText string

// 	c.OnHTML("table.wikitable:not(.sortable) tbody tr", func(row *colly.HTMLElement) {
// 		// The language header cell is a <th scope="row">.
// 		langCell := strings.TrimSpace(row.ChildText("th[scope=row]"))

// 		// If this row has a language header, update our tracking variables.
// 		if langCell != "" {
// 			code, ok := yugipediaLangMap[langCell]
// 			if !ok {
// 				// Unknown language — skip the row.
// 				currentLang = ""
// 				return
// 			}
// 			currentLang = code

// 			// The card text may span multiple rows (rowspan); grab it here
// 			// if present (it won't appear in the second row for JP/KR).
// 			currentCardText = strings.TrimSpace(row.ChildText("td:last-child"))
// 		}

// 		// Skip rows without a recognised language.
// 		if currentLang == "" {
// 			return
// 		}

// 		// The name is always in the first <td> of the row.
// 		name := strings.TrimSpace(row.ChildText("td:first-of-type"))
// 		if name == "" {
// 			// Second row of a rowspan (romanisation) — nothing new to store.
// 			return
// 		}

// 		langName, _ := cards.GetLangName(currentLang)

// 		result[currentLang] = cards.Card{
// 			Lang:        currentLang,
// 			Language:    langName,
// 			Name:        name,
// 			Description: currentCardText,
// 		}
// 	})

// 	if err := c.Visit(url); err != nil {
// 		return nil, fmt.Errorf("error scraping %s: %w", url, err)
// 	}
// 	if scrapeErr != nil {
// 		return nil, scrapeErr
// 	}

// 	return result, nil
// }
