package providers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/operaodev/cardex/internal/items"
)

type SetInfo struct {
	ExternalID     string
	SetType        string // "Structure Deck", "Booster Pack", "Collector's Set", etc.
	Medium         string // "TCG", "OCG", "TCG/OCG"
	Names          map[items.LangCode]string
	Prefixes       map[items.LangCode]string
	QuantityPerSet uint
	QuantityPerBox uint
	SetImage       string
	CardListPages  map[items.LangCode]string // lang -> URL path
}

type SetCardEntry struct {
	CardCode string
	CardName string
	Rarity   string
	Quantity uint
	Category string
	Print    string // "New", "Reprint", etc.
	IsBonus  bool
	Lang     items.LangCode
}

var (
	rePackPerBox    = regexp.MustCompile(`(\d+)\s+(?:cards?\s+)?per\s+pack`)
	reBoxPerBox     = regexp.MustCompile(`(\d+)\s+packs?\s+per\s+box`)
	rePreconstructedDeck = regexp.MustCompile(`Preconstructed\s+Deck\s+of\s+(\d+)\s+cards?`)
	reDeckCount     = regexp.MustCompile(`(\d+)\s+Decks?\s+(?:of|with)\s+(\d+)\s+cards?`)
	reTotalCards    = regexp.MustCompile(`(?:contains?|contains)\s+(\d+)\s+cards?`)
)

func parseSetPage(e *colly.HTMLElement) *SetInfo {
	set := &SetInfo{
		Names:         make(map[items.LangCode]string),
		Prefixes:      make(map[items.LangCode]string),
		CardListPages: make(map[items.LangCode]string),
	}

	set.ExternalID = strings.TrimSpace(e.DOM.Find("th.infobox-above").First().Text())

	e.DOM.Find("table.infobox tr").Each(func(_ int, row *goquery.Selection) {
		label := strings.TrimSpace(row.Find("th.infobox-label").Text())
		data := row.Find("td.infobox-data")

		switch label {
		case "Medium":
			set.Medium = strings.TrimSpace(data.Text())
		case "Type":
			set.SetType = strings.TrimSpace(data.Find("li").First().Text())
			if set.SetType == "" {
				set.SetType = strings.TrimSpace(data.Text())
			}
		case "Number of cards":
			text := data.Text()
			if num := extractNumber(text); num > 0 {
				set.QuantityPerSet = num
			}
		}

		// Names
		lang := langFromInfoboxLabel(label)
		if lang != "" {
			nameText := strings.TrimSpace(data.Text())
			set.Names[lang] = nameText
		}

		// Prefixes
		if label == "Prefix" {
			data.Find("li").Each(func(_ int, li *goquery.Selection) {
				text := li.Text()
				langCode, prefix := parsePrefix(text)
				if langCode != "" {
					set.Prefixes[langCode] = prefix
				}
			})
		}
	})

	// Parse breakdown for quantities
	set.parseBreakdown(e)

	// Parse galleries
	set.parseGalleryLinks(e)

	// Parse card list tabs
	set.parseCardListTabs(e)

	return set
}

func (s *SetInfo) parseBreakdown(e *colly.HTMLElement) {
	var breakdownText string

	e.DOM.Find("h2").Each(func(_ int, h2 *goquery.Selection) {
		if strings.Contains(h2.Text(), "Breakdown") || strings.Contains(h2.Text(), "Contents") {
			// Get all text until next h2
			siblings := h2.NextAll()
			siblings.Each(func(_ int, el *goquery.Selection) {
				if el.Is("h2") {
					return
				}
				breakdownText += " " + el.Text()
			})
		}
	})

	if breakdownText == "" {
		return
	}

	// Try to find cards per pack (booster packs)
	if m := rePackPerBox.FindStringSubmatch(breakdownText); m != nil {
		if n, err := strconv.ParseUint(m[1], 10, 32); err == nil {
			s.QuantityPerSet = uint(n)
		}
	}

	// Try to find packs per box
	if m := reBoxPerBox.FindStringSubmatch(breakdownText); m != nil {
		if n, err := strconv.ParseUint(m[1], 10, 32); err == nil {
			packs := uint(n)
			s.QuantityPerBox = s.QuantityPerSet * packs
		}
	}

	// Try to find preconstructed deck of N cards
	if m := rePreconstructedDeck.FindStringSubmatch(breakdownText); m != nil {
		if n, err := strconv.ParseUint(m[1], 10, 32); err == nil {
			s.QuantityPerSet = uint(n)
			if s.QuantityPerBox == 0 {
				s.QuantityPerBox = uint(n)
			}
		}
	}

	// Try to find "N Decks with X cards"
	if m := reDeckCount.FindStringSubmatch(breakdownText); m != nil {
		if decks, err := strconv.ParseUint(m[1], 10, 32); err == nil {
			if perDeck, err := strconv.ParseUint(m[2], 10, 32); err == nil {
				s.QuantityPerSet = uint(decks * perDeck)
				if s.QuantityPerBox == 0 {
					s.QuantityPerBox = s.QuantityPerSet
				}
			}
		}
	}
}

func (s *SetInfo) parseGalleryLinks(e *colly.HTMLElement) {
	// Get main gallery image (first gallerybox with a valid image)
	firstImg := e.DOM.Find("ul.gallery .gallerybox a.image img").First()
	if imgSrc, exists := firstImg.Attr("src"); exists {
		s.SetImage = imgSrc
	}
}

func (s *SetInfo) parseCardListTabs(e *colly.HTMLElement) {
	langFromTitle := map[string]items.LangCode{
		"English":            items.EN,
		"French":             items.FR,
		"German":             items.DE,
		"Italian":            items.IT,
		"Portuguese":         items.PT,
		"Spanish":            items.SP,
		"Japanese":           items.JP,
		"Asian-English":      items.AE,
		"Korean":             items.KR,
		"Simplified Chinese": items.SC,
		"North American English": items.EN,
	}

	e.DOM.Find("div.tabbertab").Each(func(_ int, tab *goquery.Selection) {
		title := strings.TrimSpace(tab.AttrOr("title", ""))
		lang, ok := langFromTitle[title]
		if !ok {
			return
		}

		pageName := tab.Find("div[data-page]").AttrOr("data-page", "")
		if pageName == "" {
			pageName = tab.Find(".set-list-tab").AttrOr("data-page", "")
		}
		if pageName != "" {
			s.CardListPages[lang] = pageName
		}
	})
}

func parseSetCardList(e *colly.HTMLElement, lang items.LangCode) []SetCardEntry {
	var entries []SetCardEntry

	table := e.DOM.Find("table.wikitable.sortable.card-list.set-list__main").First()
	if table.Length() == 0 {
		return entries
	}

	// Detect column indices
	colIndex := map[string]int{}
	table.Find("th").Each(func(i int, th *goquery.Selection) {
		class := th.AttrOr("class", "")
		switch {
		case strings.Contains(class, "set-list__main__header--card-number"):
			colIndex["card_number"] = i
		case strings.Contains(class, "set-list__main__header--name"):
			colIndex["name"] = i
		case strings.Contains(class, "set-list__main__header--rarity"):
			colIndex["rarity"] = i
		case strings.Contains(class, "set-list__main__header--category"):
			colIndex["category"] = i
		case strings.Contains(class, "set-list__main__header--quantity"):
			colIndex["quantity"] = i
		}
	})

	table.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
		tds := row.Find("td")
		if tds.Length() == 0 {
			return
		}

		getCell := func(key string) string {
			idx, ok := colIndex[key]
			if !ok || idx >= tds.Length() {
				return ""
			}
			return strings.TrimSpace(tds.Eq(idx).Text())
		}

		code := getCell("card_number")
		if code == "" {
			return
		}

		nameText := getCell("name")
		nameText = strings.Trim(nameText, "\"")

		rarity := getCell("rarity")
		category := getCell("category")
		printType := getCell("print")

		quantity := uint(1)
		if qtyStr := getCell("quantity"); qtyStr != "" {
			if n, err := strconv.ParseUint(qtyStr, 10, 32); err == nil {
				quantity = uint(n)
			}
		}

		entries = append(entries, SetCardEntry{
			CardCode: code,
			CardName: nameText,
			Rarity:   rarity,
			Quantity: quantity,
			Category: category,
			Print:    printType,
			Lang:     lang,
		})
	})

	return entries
}

func parseSetCardListWithBonuses(e *colly.HTMLElement, lang items.LangCode) []SetCardEntry {
	var allEntries []SetCardEntry

	// Find all set-list containers
	e.DOM.Find("div.set-list").Each(func(_ int, setList *goquery.Selection) {
		isBonus := false
		// Check if preceded by Bonus cards heading
		prev := setList.Prev()
		if prev.Length() > 0 && prev.Is("h3") && strings.Contains(prev.Text(), "Bonus") {
			isBonus = true
		}

		table := setList.Find("table.wikitable.sortable.card-list.set-list__main")
		if table.Length() == 0 {
			return
		}

		entries := parseSetCardListTable(table, lang, isBonus)
		allEntries = append(allEntries, entries...)
	})

	return allEntries
}

func parseSetCardListTable(table *goquery.Selection, lang items.LangCode, isBonus bool) []SetCardEntry {
	var entries []SetCardEntry

	colIndex := map[string]int{}
	table.Find("th").Each(func(i int, th *goquery.Selection) {
		class := th.AttrOr("class", "")
		switch {
		case strings.Contains(class, "set-list__main__header--card-number"):
			colIndex["card_number"] = i
		case strings.Contains(class, "set-list__main__header--name"):
			colIndex["name"] = i
		case strings.Contains(class, "set-list__main__header--rarity"):
			colIndex["rarity"] = i
		case strings.Contains(class, "set-list__main__header--category"):
			colIndex["category"] = i
		case strings.Contains(class, "set-list__main__header--print"):
			colIndex["print"] = i
		case strings.Contains(class, "set-list__main__header--quantity"):
			colIndex["quantity"] = i
		}
	})

	table.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
		tds := row.Find("td")
		if tds.Length() == 0 {
			return
		}

		getCell := func(key string) string {
			idx, ok := colIndex[key]
			if !ok || idx >= tds.Length() {
				return ""
			}
			return strings.TrimSpace(tds.Eq(idx).Text())
		}

		code := getCell("card_number")
		if code == "" {
			return
		}

		nameText := getCell("name")
		nameText = strings.Trim(nameText, "\"")

		rarity := getCell("rarity")
		category := getCell("category")
		printType := getCell("print")

		quantity := uint(1)
		if qtyStr := getCell("quantity"); qtyStr != "" {
			if n, err := strconv.ParseUint(qtyStr, 10, 32); err == nil {
				quantity = uint(n)
			}
		}

		entries = append(entries, SetCardEntry{
			CardCode: code,
			CardName: nameText,
			Rarity:   rarity,
			Quantity: quantity,
			Category: category,
			Print:    printType,
			IsBonus:  isBonus,
			Lang:     lang,
		})
	})

	return entries
}

func langFromInfoboxLabel(label string) items.LangCode {
	switch label {
	case "English":
		return items.EN
	case "French":
		return items.FR
	case "German":
		return items.DE
	case "Italian":
		return items.IT
	case "Portuguese":
		return items.PT
	case "Spanish":
		return items.SP
	case "Japanese":
		return items.JP
	case "Korean":
		return items.KR
	case "Simplified Chinese":
		return items.SC
	}
	return ""
}

func parsePrefix(text string) (items.LangCode, string) {
	// Format: "SR14-EN (en)" or "CH01-FR (fr)"
	text = strings.TrimSpace(text)

	// Find language code in parentheses
	langMatch := regexp.MustCompile(`\(([a-z]{2})\)$`)
	m := langMatch.FindStringSubmatch(text)
	if m == nil {
		return "", ""
	}

	langCode := langCodeFromISO(m[1])
	prefix := strings.TrimSpace(strings.TrimSuffix(text, m[0]))

	return langCode, prefix
}

func langCodeFromISO(iso string) items.LangCode {
	switch iso {
	case "en":
		return items.EN
	case "fr":
		return items.FR
	case "de":
		return items.DE
	case "it":
		return items.IT
	case "pt":
		return items.PT
	case "sp":
		return items.SP
	case "jp":
		return items.JP
	case "kr":
		return items.KR
	case "sc":
		return items.SC
	}
	return ""
}

func extractNumber(text string) uint {
	re := regexp.MustCompile(`\b(\d+)\b`)
	m := re.FindStringSubmatch(text)
	if m == nil {
		return 0
	}
	if n, err := strconv.ParseUint(m[1], 10, 32); err == nil {
		return uint(n)
	}
	return 0
}

func setCardsURL(baseURL, pageName string) string {
	return fmt.Sprintf("%s/wiki/%s", baseURL, strings.ReplaceAll(pageName, " ", "_"))
}

func convertSetToItems(set *SetInfo, cardEntries map[items.LangCode][]SetCardEntry) []items.Item {
	var result []items.Item

	for lang, cards := range cardEntries {
		for _, card := range cards {
			name := set.Names[lang]
			if name == "" {
				name = set.Names[items.EN]
			}
			if name == "" {
				name = set.ExternalID
			}

			item := items.Item{
				Type:           items.ItemTypeSet,
				ExternalID:     set.ExternalID,
				SetExternalID:  set.ExternalID,
				TCG:            items.YGO,
				Lang:           lang,
				Name:           name,
				SetName:        name,
				SetCode:        card.CardCode,
				CardTypes:      card.Category,
				SetType:        set.SetType,
				QuantityPerSet: card.Quantity,
				QuantityPerBox: set.QuantityPerBox,
			}

			if set.SetImage != "" {
				item.SetImage = set.SetImage
			}

			result = append(result, item)
		}
	}

	return result
}
