package providers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gocolly/colly"
	"github.com/operaodev/cardex/internal/items"
)

const (
	scrapeBatchSize      = 300
	scrapeBatchPause     = 3 * time.Second
	scrapeParallelism    = 4
	scrapeDelay          = 1 * time.Second
	scrapeRequestTimeout = 50 * time.Second
	httpClientTimeout    = 50 * time.Second
	progressLogInterval  = 100
)

type YGOCard struct {
	ExternalID     string
	ID             uint              `json:"id"`
	Name           string            `json:"name"`
	Types          string            `json:"humanReadableCardType"`
	Description    string            `json:"desc"`
	Archetype      string            `json:"archetype"`
	Images         []items.CardImage `json:"card_images"`
	Lang           items.LangCode    `json:"lang"`
	QuantityPerSet int               `json:"quantity_per_set"`
}

type YGOSet struct {
	SetExternalID string
	Lang          items.LangCode
	SetName       string
	SetCode       string
	SetRegionCode string
	Description   string
	SetType       string
}

// UniqueKey devuelve la clave de mapa para un YGOCard: "ExternalID-Lang"
func (y YGOCard) UniqueKey() string {
	return fmt.Sprintf("%s-%s", y.ExternalID, y.Lang)
}

type YGOProvider struct {
	httpClient       *http.Client
	ygoproBaseUrl    string
	yugipediaBaseUrl string
}

func NewYGOProvider() *YGOProvider {
	return &YGOProvider{
		httpClient:       &http.Client{Timeout: httpClientTimeout},
		ygoproBaseUrl:    "https://db.ygoprodeck.com/api/v7",
		yugipediaBaseUrl: "https://yugipedia.com",
	}
}

func (y *YGOProvider) FetchCards() ([]items.Item, error) {
	englishCards, err := y.fetchCardsYGOPro("")
	if err != nil {
		return nil, err
	}
	return y.scrapeCards(englishCards)
}

func (y *YGOProvider) FetchCardsByName(name string) ([]items.Item, error) {
	englishCards, err := y.fetchCardsYGOPro(name)
	if err != nil {
		return nil, err
	}
	return y.scrapeCards(englishCards)
}

// wikiPagePath convierte un nombre de carta en un segmento de ruta URL compatible con MediaWiki.
// Los espacios se convierten en guiones bajos; solo ? y # se codifican en porcentaje porque
// se interpretarían como delimitadores de cadena de consulta / fragmento. Todos los demás caracteres
// especiales (comas, barras, signos de exclamación, etc.) se dejan tal cual porque
// MediaWiki los maneja nativamente.
func wikiPagePath(name string) string {
	p := strings.ReplaceAll(name, " ", "_")
	p = strings.ReplaceAll(p, "?", "%3F")
	p = strings.ReplaceAll(p, "#", "%23")
	return p
}

func createColly() *colly.Collector {
	c := colly.NewCollector(
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"),
	)
	c.SetRequestTimeout(scrapeRequestTimeout)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: scrapeParallelism,
		Delay:       scrapeDelay,
	})

	setBrowserHeaders := func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Cache-Control", "max-age=0")
		r.Headers.Set("Sec-Ch-Ua", `"Not/A)Brand";v="8", "Chromium";v="125", "Google Chrome";v="125"`)
		r.Headers.Set("Sec-Ch-Ua-Mobile", "?0")
		r.Headers.Set("Sec-Ch-Ua-Platform", `"Linux"`)
		r.Headers.Set("Sec-Fetch-Dest", "document")
		r.Headers.Set("Sec-Fetch-Mode", "navigate")
		r.Headers.Set("Sec-Fetch-Site", "none")
		r.Headers.Set("Sec-Fetch-User", "?1")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
	}
	c.OnRequest(setBrowserHeaders)
	return c
}

// scrapeCards visita la página de Yugipedia de cada carta, analiza traducciones y entradas de impresión,
// y devuelve la sección de elementos ensamblada. Las cartas se procesan por lotes
// para mantener el uso de memoria acotado y dar pausas de enfriamiento entre lotes
// para Cloudflare, reduciendo la posibilidad de ser limitado en velocidad.
func (y *YGOProvider) scrapeCards(englishCards []YGOCard) ([]items.Item, error) {
	log.Printf("[scrapeCards] iniciando escaneo para %d carta(s)", len(englishCards))

	enCardByName := make(map[string]YGOCard, len(englishCards))
	for _, card := range englishCards {
		enCardByName[card.Name] = card
	}

	c := createColly()

	var (
		mu              sync.Mutex
		allItems        []items.Item
		cardsWithPrints = make(map[string]bool)
		galleryURLs     = make(map[string]string) // ExternalID → relative gallery URL
		processedCount  atomic.Int64
		cfBlockCount    atomic.Int64
	)

	c.OnError(func(r *colly.Response, err error) {
		cardName := r.Request.Ctx.Get("name")
		if r != nil && (r.StatusCode == http.StatusForbidden || r.StatusCode == http.StatusServiceUnavailable) {
			blocks := cfBlockCount.Add(1)
			log.Printf("[scrapeCards] BLOQUEO CLOUDFLARE #%d para carta %q (%s): estado %d",
				blocks, cardName, r.Request.URL, r.StatusCode)
		} else {
			log.Printf("[scrapeCards] error HTTP para carta %q (%s): %v",
				cardName, r.Request.URL, err)
		}
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {
		cardName := e.Request.Ctx.Get("name")

		count := processedCount.Add(1)
		if count%progressLogInterval == 0 {
			log.Printf("[scrapeCards] progreso: %d/%d cartas procesadas, %d elementos hasta ahora",
				count, len(englishCards), len(allItems))
		}

		mu.Lock()
		enCard, ok := enCardByName[cardName]
		mu.Unlock()
		if !ok {
			log.Printf("[scrapeCards] carta %q no encontrada en el mapa de búsqueda (posible redirección o discrepancia de nombre)", cardName)
			return
		}

		translatedCards := parseCards(e, enCard)
		translatedCards[enCard.UniqueKey()] = enCard

		localItems := mapPrints(e, enCard, translatedCards)

		// Extract gallery URL from the card page
		if gURL := parseGalleryLink(e); gURL != "" {
			mu.Lock()
			galleryURLs[enCard.ExternalID] = gURL
			mu.Unlock()
		}

		mu.Lock()
		if len(localItems) > 0 {
			cardsWithPrints[cardName] = true
			allItems = append(allItems, localItems...)
		} else {
			log.Printf("[scrapeCards] ningún elemento de impresión encontrado para la carta %q — verifique las tablas CTS en %s", cardName, e.Request.URL)
		}
		mu.Unlock()
	})

	for i := 0; i < len(englishCards); i += scrapeBatchSize {
		end := min(i+scrapeBatchSize, len(englishCards))

		for _, card := range englishCards[i:end] {
			ctx := colly.NewContext()
			ctx.Put("name", card.Name)
			id := fmt.Sprintf("%08d", card.ID)
			wikiURL := fmt.Sprintf("%s/wiki/%s", y.yugipediaBaseUrl, id)
			_ = c.Request("GET", wikiURL, nil, ctx, nil)
		}

		c.Wait()

		if end < len(englishCards) {
			log.Printf("[scrapeCards] lote completo: %d/%d cartas procesadas, pausando %v",
				end, len(englishCards), scrapeBatchPause)
			time.Sleep(scrapeBatchPause)
		}
	}

	blocks := cfBlockCount.Load()
	if blocks > 0 {
		log.Printf("[scrapeCards] ADVERTENCIA: %d bloqueos de Cloudflare detectados durante el escaneo", blocks)
	}
	log.Printf("[scrapeCards] terminado — total de elementos ensamblados: %d (de %d cartas), %d galerías encontradas", len(allItems), len(englishCards), len(galleryURLs))

	// Paso 3: Escanear galerías para cartas con impresiones
	allItems = y.scrapeGallery(allItems, galleryURLs)

	return allItems, nil
}

// scrapeGallery visita las páginas de Galería de Cartas para cartas que tienen impresiones y enriquece
// los elementos con PrintURLSmall, PrintURLLarge, Edición y Código de Rareza.
func (y *YGOProvider) scrapeGallery(allItems []items.Item, galleryURLs map[string]string) []items.Item {
	if len(galleryURLs) == 0 {
		return allItems
	}

	log.Printf("[scrapeGallery] iniciando escaneo de galería para %d carta(s) con galería", len(galleryURLs))

	// Construir índice de elementos: SetExternalID|Code|Lang|Rarity|Edition → *Item
	itemIndex := make(map[string]*items.Item, len(allItems))
	for i := range allItems {
		key := fmt.Sprintf("%s|%s|%s|%s|%s", allItems[i].SetExternalID, allItems[i].Code, allItems[i].Lang, allItems[i].Rarity, allItems[i].Edition)
		itemIndex[key] = &allItems[i]
	}

	log.Printf("[scrapeGallery] índice de elementos construido con %d entradas", len(itemIndex))

	gc := colly.NewCollector(
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"),
	)
	gc.SetRequestTimeout(scrapeRequestTimeout)
	gc.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: scrapeParallelism,
		Delay:       scrapeDelay,
	})

	setBrowserHeaders := func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Cache-Control", "max-age=0")
		r.Headers.Set("Sec-Ch-Ua", `"Not/A)Brand";v="8", "Chromium";v="125", "Google Chrome";v="125"`)
		r.Headers.Set("Sec-Ch-Ua-Mobile", "?0")
		r.Headers.Set("Sec-Ch-Ua-Platform", `"Linux"`)
		r.Headers.Set("Sec-Fetch-Dest", "document")
		r.Headers.Set("Sec-Fetch-Mode", "navigate")
		r.Headers.Set("Sec-Fetch-Site", "none")
		r.Headers.Set("Sec-Fetch-User", "?1")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
	}
	gc.OnRequest(setBrowserHeaders)

	var (
		mu             sync.Mutex
		processedCount atomic.Int64
		matchedCount   atomic.Int64
	)

	gc.OnError(func(r *colly.Response, err error) {
		externalID := r.Request.Ctx.Get("externalID")
		if r != nil && r.StatusCode == http.StatusNotFound {
			// La página de galería no existe, omitir silenciosamente
			return
		}
		log.Printf("[scrapeGallery] error HTTP para carta %q (%s): %v",
			externalID, r.Request.URL, err)
	})

	gc.OnHTML("body", func(e *colly.HTMLElement) {
		count := processedCount.Add(1)
		if count%progressLogInterval == 0 {
			log.Printf("[scrapeGallery] progreso: %d/%d galerías procesadas, %d coincidencias hasta ahora",
				count, len(galleryURLs), matchedCount.Load())
		}

		entries := parseGallery(e)

		mu.Lock()
		for _, entry := range entries {
			// Intentar coincidencia exacta con Edition primero
			key := fmt.Sprintf("%s|%s|%s|%s|%s", entry.Set, entry.Code, entry.Lang, entry.Rarity, entry.Edition)
			item, ok := itemIndex[key]
			if !ok {
				// Fallback: coincidir sin Edition (para elementos de CTS que nunca tienen Edition)
				fallbackKey := fmt.Sprintf("%s|%s|%s|%s|", entry.Set, entry.Code, entry.Lang, entry.Rarity)
				item, ok = itemIndex[fallbackKey]
			}
			if ok {
				item.PrintURLSmall = entry.PrintURLSmall
				item.PrintURLLarge = entry.PrintURLLarge
				item.Edition = entry.Edition
				item.RarityCode = entry.RarityCode
				matchedCount.Add(1)
			} else {
				log.Printf("[scrapeGallery] ninguna coincidencia para la entrada de galería: Set=%q Código=%q Lang=%s Rareza=%q Edición=%q", entry.Set, entry.Code, entry.Lang, entry.Rarity, entry.Edition)
			}
		}
		mu.Unlock()
	})

	// Convertir mapa a slice para procesamiento por lotes
	type galleryEntry struct {
		externalID string
		galleryURL string
	}
	var galleries []galleryEntry
	for externalID, gURL := range galleryURLs {
		galleries = append(galleries, galleryEntry{externalID: externalID, galleryURL: gURL})
	}

	for i := 0; i < len(galleries); i += scrapeBatchSize {
		end := min(i+scrapeBatchSize, len(galleries))

		for _, g := range galleries[i:end] {
			ctx := colly.NewContext()
			ctx.Put("externalID", g.externalID)
			galleryURL := fmt.Sprintf("%s%s", y.yugipediaBaseUrl, g.galleryURL)
			_ = gc.Request("GET", galleryURL, nil, ctx, nil)
		}

		gc.Wait()

		if end < len(galleries) {
			log.Printf("[scrapeGallery] pausa de lote: %d/%d galerías realizadas, pausando %v",
				end, len(galleries), scrapeBatchPause)
			time.Sleep(scrapeBatchPause)
		}
	}

	log.Printf("[scrapeGallery] terminado — coincidieron %d entradas de galería con elementos", matchedCount.Load())
	return allItems
}

// fetchCardsYGOPro llama a la API de YGOPRODeck y devuelve las cartas en inglés.
func (y *YGOProvider) fetchCardsYGOPro(name string) ([]YGOCard, error) {
	var reqURL string
	if name != "" {
		reqURL = fmt.Sprintf("%s/cardinfo.php?fname=%s", y.ygoproBaseUrl, url.QueryEscape(name))
	} else {
		reqURL = fmt.Sprintf("%s/cardinfo.php", y.ygoproBaseUrl)
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := y.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected code YGOPRODeck: %d", resp.StatusCode)
	}

	var response struct {
		Data []YGOCard `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	if len(response.Data) == 0 {
		return nil, fmt.Errorf("no cards found")
	}

	cards := make([]YGOCard, 0, len(response.Data))
	for _, card := range response.Data {
		card.ExternalID = card.Name
		card.Lang = items.EN
		card.Types = strings.ReplaceAll(card.Types, " ", "/")
		cards = append(cards, card)
	}
	return cards, nil
}

// FetchSets escanea páginas de conjuntos para enriquecer elementos existentes con información del set.
func (y *YGOProvider) FetchSets(existingItems []items.Item) ([]items.Item, error) {
	log.Printf("[FetchSets] iniciando escaneo de conjuntos con %d elementos existentes", len(existingItems))

	// 1. Construir índice de conjuntos a partir de elementos existentes (map setExternalID -> items)
	setIndex := make(map[string][]*items.Item)
	for i := range existingItems {
		if existingItems[i].Type == items.ItemTypeCard && existingItems[i].SetExternalID != "" {
			setIndex[existingItems[i].SetExternalID] = append(setIndex[existingItems[i].SetExternalID], &existingItems[i])
		}
	}

	setNames := make([]string, 0, len(setIndex))
	for name := range setIndex {
		setNames = append(setNames, name)
	}

	log.Printf("[FetchSets] encontrados %d conjuntos únicos para escanear", len(setNames))

	// 2. Escanear páginas principales de conjuntos
	sc := colly.NewCollector(
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"),
	)
	sc.SetRequestTimeout(scrapeRequestTimeout)
	sc.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: scrapeParallelism,
		Delay:       scrapeDelay,
	})

	setBrowserHeaders := func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Cache-Control", "max-age=0")
		r.Headers.Set("Sec-Ch-Ua", `"Not/A)Brand";v="8", "Chromium";v="125", "Google Chrome";v="125"`)
		r.Headers.Set("Sec-Ch-Ua-Mobile", "?0")
		r.Headers.Set("Sec-Ch-Ua-Platform", `"Linux"`)
		r.Headers.Set("Sec-Fetch-Dest", "document")
		r.Headers.Set("Sec-Fetch-Mode", "navigate")
		r.Headers.Set("Sec-Fetch-Site", "none")
		r.Headers.Set("Sec-Fetch-User", "?1")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
	}
	sc.OnRequest(setBrowserHeaders)

	var (
		mu             sync.Mutex
		allItems       []items.Item
		processedCount atomic.Int64
	)

	// Escanear páginas de lista de cartas para cada idioma
	lc := colly.NewCollector(
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"),
	)
	lc.SetRequestTimeout(scrapeRequestTimeout)
	lc.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: scrapeParallelism,
		Delay:       scrapeDelay,
	})
	lc.OnRequest(setBrowserHeaders)

	lc.OnError(func(r *colly.Response, err error) {
		setName := r.Request.Ctx.Get("setName")
		log.Printf("[FetchSets] [lista] error HTTP para conjunto %q lang %q (%s): %v",
			setName, r.Request.Ctx.Get("lang"), r.Request.URL, err)
	})

	lc.OnHTML("body", func(e *colly.HTMLElement) {
		setName := e.Request.Ctx.Get("setName")
		langStr := e.Request.Ctx.Get("lang")
		lang := items.LangCode(langStr)

		entries := parseSetCardListWithBonuses(e, lang)

		mu.Lock()
		cardsForSet := setIndex[setName]
		mu.Unlock()

		for _, entry := range entries {
			for _, item := range cardsForSet {
				if item.SetCode == entry.CardCode && item.Lang == lang {
					item.QuantityPerSet = entry.Quantity
				}
			}
		}
	})

	sc.OnError(func(r *colly.Response, err error) {
		cardName := r.Request.Ctx.Get("name")
		if r != nil && r.StatusCode == http.StatusNotFound {
			log.Printf("[FetchSets] página de conjunto no encontrada: %q (%s)", cardName, r.Request.URL)
			return
		}
		log.Printf("[FetchSets] error HTTP para conjunto %q (%s): %v",
			cardName, r.Request.URL, err)
	})

	sc.OnHTML("body", func(e *colly.HTMLElement) {
		setName := e.Request.Ctx.Get("name")
		count := processedCount.Add(1)
		if count%progressLogInterval == 0 {
			log.Printf("[FetchSets] progreso: %d/%d conjuntos procesados", count, len(setNames))
		}

		// Analizar página principal del conjunto
		setInfo := parseSetPage(e)
		if setInfo.ExternalID == "" {
			setInfo.ExternalID = setName
		}

		mu.Lock()
		cardsForSet := setIndex[setName]
		mu.Unlock()

		if len(cardsForSet) == 0 {
			return
		}

		// Procesar las pestañas de lista de cartas de cada idioma
		for lang, pageName := range setInfo.CardListPages {
			// Omitir si no hay elementos para este idioma
			hasLang := false
			for _, item := range cardsForSet {
				if item.Lang == lang {
					hasLang = true
					break
				}
			}
			if !hasLang {
				continue
			}

			listURL := setCardsURL(y.yugipediaBaseUrl, pageName)
			listCtx := colly.NewContext()
			listCtx.Put("setName", setName)
			listCtx.Put("lang", string(lang))
			_ = lc.Request("GET", listURL, nil, listCtx, nil)
		}

		// Enriquecer elementos existentes con información del conjunto
		for _, item := range cardsForSet {
			item.SetType = setInfo.SetType
			item.QuantityPerBox = setInfo.QuantityPerBox
			if setInfo.SetImage != "" && item.SetImage == "" {
				item.SetImage = setInfo.SetImage
			}
		}
	})

	// 3. Escanear páginas principales de conjuntos en lotes
	for i := 0; i < len(setNames); i += scrapeBatchSize {
		end := min(i+scrapeBatchSize, len(setNames))

		for _, name := range setNames[i:end] {
			ctx := colly.NewContext()
			ctx.Put("name", name)
			wikiURL := fmt.Sprintf("%s/wiki/%s", y.yugipediaBaseUrl, wikiPagePath(name))
			_ = sc.Request("GET", wikiURL, nil, ctx, nil)
		}

		sc.Wait()

		if end < len(setNames) {
			log.Printf("[FetchSets] pausa de lote: %d/%d conjuntos realizados, pausando %v",
				end, len(setNames), scrapeBatchPause)
			time.Sleep(scrapeBatchPause)
		}
	}

	lc.Wait()

	log.Printf("[FetchSets] terminado — enriquecidos %d elementos con datos de conjunto", len(allItems))
	return existingItems, nil
}
