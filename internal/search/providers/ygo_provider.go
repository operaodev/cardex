package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

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

func (p *YGOProvider) FetchAllCards() ([]search.ResultCard, error) {
	reqURL := fmt.Sprintf("%s/cardinfo.php?", p.ygoproBaseURL)
	ygoproCards, err := p.fetchFromYGOPro(reqURL)
	if err != nil {
		return nil, err
	}

	if len(ygoproCards) == 0 {
		return []search.ResultCard{}, nil
	}

	// Dividir las ~11k cartas en lotes de 20 (límite de Yugipedia)
	const (
		batchSize     = 20
		maxConcurrent = 10 // goroutines concurrentes máximas hacia Yugipedia
	)

	var batches [][]YGOProCard
	for i := 0; i < len(ygoproCards); i += batchSize {
		end := min(i+batchSize, len(ygoproCards))
		batches = append(batches, ygoproCards[i:end])
	}

	type batchResult struct {
		index   int
		results []search.ResultCard
		err     error
	}

	// context para cancelar goroutines pendientes si ocurre un error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sem := make(chan struct{}, maxConcurrent)
	resultsCh := make(chan batchResult, len(batches))

	var wg sync.WaitGroup
	for i, batch := range batches {
		wg.Add(1)
		go func(idx int, b []YGOProCard) {
			defer wg.Done()

			// Adquirir slot del semaforo o cancelar si hubo error en otro batch
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				resultsCh <- batchResult{index: idx, err: ctx.Err()}
				return
			}

			names := make([]string, len(b))
			for j, c := range b {
				names[j] = c.Name
			}

			yugipediaResults, fetchErr := p.getTotalCardInfo(names...)
			if fetchErr != nil {
				resultsCh <- batchResult{index: idx, err: fetchErr}
				return
			}

			yugipediaByName := make(map[string]search.ResultCard, len(yugipediaResults))
			for _, yResult := range yugipediaResults {
				if enName, ok := yResult.Names[cards.EN]; ok {
					yugipediaByName[enName] = yResult
				}
			}

			batchCards := make([]search.ResultCard, 0, len(b))
			for _, ygoproCard := range b {
				yResult := yugipediaByName[ygoproCard.Name]
				batchCards = append(batchCards, p.mergeWithYGOPro(ygoproCard, yResult))
			}

			resultsCh <- batchResult{index: idx, results: batchCards}
		}(i, batch)
	}

	// Cerrar el canal cuando todas las goroutines terminen
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Recolectar resultados en orden usando el index del batch
	orderedResults := make([][]search.ResultCard, len(batches))
	var firstErr error
	for br := range resultsCh {
		if br.err != nil && firstErr == nil {
			firstErr = br.err
			cancel() // abortar goroutines pendientes
		}
		if br.results != nil {
			orderedResults[br.index] = br.results
		}
	}

	if firstErr != nil {
		return nil, fmt.Errorf("error procesando lote desde Yugipedia: %w", firstErr)
	}

	// Aplanar los resultados en orden de los batches
	allResults := make([]search.ResultCard, 0, len(ygoproCards))
	for _, batch := range orderedResults {
		allResults = append(allResults, batch...)
	}

	return allResults, nil
}

func (p *YGOProvider) FetchCardByID(id string) (search.ResultCard, error) {
	reqURL := fmt.Sprintf("%s/cardinfo.php?id=%s", p.ygoproBaseURL, id)
	cards, err := p.fetchFromYGOPro(reqURL)
	if err != nil {
		return search.ResultCard{}, err
	}

	ygoproResult := cards[0]
	yugipediaResults, err := p.getTotalCardInfo(ygoproResult.Name)
	if err != nil {
		return search.ResultCard{}, fmt.Errorf("error obteniendo info de Yugipedia: %w", err)
	}

	var yResult search.ResultCard
	if len(yugipediaResults) > 0 {
		yResult = yugipediaResults[0]
	}

	return p.mergeWithYGOPro(ygoproResult, yResult), nil
}

func (p *YGOProvider) FetchCardsByName(name string) ([]search.ResultCard, error) {
	reqURL := fmt.Sprintf("%s/cardinfo.php?fname=%s", p.ygoproBaseURL, url.QueryEscape(name))
	ygoproCards, err := p.fetchFromYGOPro(reqURL)
	if err != nil {
		return nil, err
	}

	if len(ygoproCards) == 0 {
		return []search.ResultCard{}, nil
	}

	// Limitar a 20 para no saturar la API de Yugipedia
	limit := min(len(ygoproCards), 20)
	ygoproCards = ygoproCards[:limit]

	allNames := make([]string, limit)
	for i, c := range ygoproCards {
		allNames[i] = c.Name
	}

	yugipediaResults, err := p.getTotalCardInfo(allNames...)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo info de Yugipedia: %w", err)
	}

	yugipediaByName := make(map[string]search.ResultCard, len(yugipediaResults))
	for _, yResult := range yugipediaResults {
		if enName, ok := yResult.Names[cards.EN]; ok {
			yugipediaByName[enName] = yResult
		}
	}

	results := make([]search.ResultCard, 0, len(ygoproCards))
	for _, ygoproCard := range ygoproCards {
		yResult := yugipediaByName[ygoproCard.Name]
		results = append(results, p.mergeWithYGOPro(ygoproCard, yResult))
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

func (p *YGOProvider) fetchFromYGOPro(url string) ([]YGOProCard, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
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

	var response struct {
		Data []YGOProCard `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (p *YGOProvider) mergeWithYGOPro(ygoproCard YGOProCard, yResult search.ResultCard) search.ResultCard {
	result := yResult

	result.ExternalID = fmt.Sprintf("%d", ygoproCard.ID)

	types := strings.Split(ygoproCard.Types, " ")
	if len(types) > 0 {
		result.Type = types[len(types)-1]
		result.Subtypes = types[:len(types)-1]
	}

	result.Archetype = ygoproCard.Archetype
	result.Images = ygoproCard.Images
	result.TCG = cards.TCGYugioh

	result.Sources = append(result.Sources, ygoproCard.YgoprodeckURL)
	result.Sources = append(result.Sources, "https://yugipedia.com/wiki/"+url.QueryEscape(ygoproCard.Name))

	if result.Names == nil {
		result.Names = make(map[cards.LangCode]string)
	}
	result.Names[cards.EN] = ygoproCard.Name

	return result
}
