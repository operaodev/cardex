package providers

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/operaodev/cardex/internal/items"
)

type translations map[items.LangCode]string

func serieCode(code string) string {
	parts := strings.Split(code, "-")
	if len(parts) != 2 {
		return code
	}

	re := regexp.MustCompile(`\d+`)
	number := re.FindString(parts[1])

	return parts[0] + "-" + number
}

func regionCode(setCode string) string {
	re := regexp.MustCompile(`[0-9]+`)
	return re.ReplaceAllString(setCode, "")
}

func setCode(code string) string {
	before, _, _ := strings.Cut(code, "-")
	return before
}

// parseCards scrapes the Yugipedia "Other languages" wikitable for name/description
// translations of the given card and returns a map keyed by "ExternalID-Lang".
// It specifically targets the table whose header row contains a "Language" column
// to avoid picking up unrelated wikitables (limitation history, in other media, etc.).
func parseCards(e *colly.HTMLElement, card YGOCard) map[string]YGOCard {
	names := translations{}
	descs := translations{}

	e.ForEach("table.wikitable", func(_ int, table *colly.HTMLElement) {
		isLangTable := false
		table.ForEach("tr th[scope=col]", func(_ int, th *colly.HTMLElement) {
			if strings.TrimSpace(th.Text) == "Language" {
				isLangTable = true
			}
		})
		if !isLangTable {
			return
		}

		table.ForEach("tbody tr", func(_ int, row *colly.HTMLElement) {
			langCell := strings.TrimSpace(row.ChildText("th[scope=row]"))
			if langCell == "" {
				return
			}

			tds := row.DOM.Find("td")
			if tds.Length() < 1 {
				return
			}

			nameVal := strings.TrimSpace(tds.First().Text())

			descVal := ""
			if tds.Length() >= 2 {
				descTD := tds.Eq(1)
				descTD.Find("br").Each(func(_ int, br *goquery.Selection) {
					br.ReplaceWithHtml("\n")
				})
				descVal = strings.TrimSpace(descTD.Text())
			}

			if nameVal != "" {
				filterLangElement(langCell, nameVal, names)
			}
			if descVal != "" {
				filterLangElement(langCell, descVal, descs)
			}
		})
	})

	result := make(map[string]YGOCard, len(names))
	for lang, name := range names {
		newCard := YGOCard{
			ExternalID:  card.ExternalID,
			Name:        name,
			Lang:        lang,
			Description: descs[lang],
			Types:       card.Types,
			Archetype:   card.Archetype,
			Images:      card.Images,
		}
		result[newCard.UniqueKey()] = newCard
	}
	return result
}

// mapPrints builds Item prints from the CTS tables on the page, resolving each
// print's language against translatedCards (keyed by "ExternalID-Lang").
func mapPrints(e *colly.HTMLElement, card YGOCard, translatedCards map[string]YGOCard) []items.Item {
	printEntries := parseCTSTables(e)

	var result []items.Item
	for _, entry := range printEntries {
		regionCardKey := card.ExternalID + "-" + string(entry.Lang)
		regionCard, exists := translatedCards[regionCardKey]
		if !exists {
			// Fallback: use the EN card data when no translation is available
			// (e.g., JP/KR/AE/TC names are in the infobox, not in the wikitable).
			enKey := card.ExternalID + "-" + string(items.EN)
			regionCard, exists = translatedCards[enKey]
			if !exists {
				continue
			}
		}

		for _, rarity := range entry.Rarities {
			item := items.Item{
				Type:          items.ItemTypeCard,
				ExternalID:    card.ExternalID,
				SetExternalID: entry.SetExternalID,
				TCG:           items.YGO,
				Code:          entry.Code,
				Lang:          entry.Lang,
				Rarity:        rarity,
				Name:          regionCard.Name,
				SetName:       entry.SetName,
				SetCode:       setCode(entry.Code),
				SerieCode:     serieCode(entry.Code),
				SetRegionCode: regionCode(entry.Code),
				Description:   regionCard.Description,
				CardTypes:     regionCard.Types,
				Archetype:     regionCard.Archetype,
				Images:        regionCard.Images,
			}
			result = append(result, item)
		}
	}
	return result
}

// printEntry groups the data from a single row of a CTS table.
type printEntry struct {
	SetExternalID string
	SetName       string
	Code          string
	Rarities      []string
	Lang          items.LangCode
}

func parseCTSTables(e *colly.HTMLElement) []printEntry {
	tableToLang := map[string]items.LangCode{
		"cts--EN": items.EN,
		"cts--NA": items.EN, // North America → EN
		"cts--EU": items.EN, // Europe → EN
		"cts--OC": items.EN, // Oceania → EN
		"cts--FR": items.FR,
		"cts--FC": items.FR, // French-Canada → FR
		"cts--DE": items.DE,
		"cts--IT": items.IT,
		"cts--PT": items.PT,
		"cts--SP": items.SP,
		"cts--JP": items.JP,
		"cts--JA": items.JP, // Japan-Asia → JP
		"cts--AE": items.AE,
		"cts--KR": items.KR,
		"cts--TC": items.TC,
		"cts--SC": items.SC,
	}

	var entries []printEntry

	e.ForEach("table.cts", func(_ int, table *colly.HTMLElement) {
		tableID, exists := table.DOM.Attr("id")
		if !exists {
			return
		}
		lang, ok := tableToLang[tableID]
		if !ok {
			return
		}

		// Detect column indices from the header row.
		// Yugipedia renders CTS tables without a <thead>: the header <tr> with <th>
		// cells sits directly inside <tbody>, so we use "tr th" instead of "thead th".
		colIndex := map[string]int{}
		table.ForEach("tr th", func(i int, th *colly.HTMLElement) {
			class, _ := th.DOM.Attr("class")
			switch {
			case strings.Contains(class, "cts__header--number"):
				colIndex["number"] = i
			case strings.Contains(class, "cts__header--set-localized"):
				colIndex["setlocal"] = i
			case strings.Contains(class, "cts__header--set"):
				colIndex["set"] = i
			case strings.Contains(class, "cts__header--rarity"):
				colIndex["rarity"] = i
			}
		})

		table.ForEach("tbody tr", func(_ int, row *colly.HTMLElement) {
			tds := row.DOM.Find("td")
			if tds.Length() == 0 {
				return
			}

			get := func(key string) string {
				idx, ok := colIndex[key]
				if !ok || idx >= tds.Length() {
					return ""
				}
				return strings.TrimSpace(tds.Eq(idx).Text())
			}

			code := get("number")
			// SetExternalID is always the English set name.
			setExternalID := get("set")
			// SetName is the localized name when available.
			setName := setExternalID
			if local := get("setlocal"); local != "" {
				setName = local
			}

			// Rarities may be multiple <a> tags inside a single cell.
			var rarities []string
			if idx, ok := colIndex["rarity"]; ok {
				cell := tds.Eq(idx)
				cell.Find("a").Each(func(_ int, a *goquery.Selection) {
					if r := strings.TrimSpace(a.Text()); r != "" {
						rarities = append(rarities, r)
					}
				})
				// Fallback: use plain text when rarity is not wrapped in <a>.
				if len(rarities) == 0 {
					if r := strings.TrimSpace(cell.Text()); r != "" {
						rarities = append(rarities, r)
					}
				}
			}

			if len(rarities) == 0 {
				return
			}

			entries = append(entries, printEntry{
				SetExternalID: setExternalID,
				SetName:       setName,
				Code:          code,
				Rarities:      rarities,
				Lang:          lang,
			})
		})
	})

	return entries
}

// filterLangElement maps a Yugipedia language label to a LangCode and stores
// the value in the provided map.
func filterLangElement(key string, value string, m map[items.LangCode]string) {
	switch key {
	case "English":
		m[items.EN] = value
	case "French":
		m[items.FR] = value
	case "German":
		m[items.DE] = value
	case "Italian":
		m[items.IT] = value
	case "Portuguese":
		m[items.PT] = value
	case "Spanish":
		m[items.SP] = value
	case "Japanese":
		m[items.JP] = value
	case "Asian-English":
		m[items.AE] = value
	case "Korean":
		m[items.KR] = value
	case "Traditional Chinese":
		m[items.TC] = value
	case "Simplified Chinese":
		m[items.SC] = value
	}
}

// FilterLangElement is the exported alias kept for backward compatibility.
func FilterLangElement(key, value string, m map[items.LangCode]string) {
	filterLangElement(key, value, m)
}

// galleryEntry represents a single image entry from a card gallery page.
type galleryEntry struct {
	Code          string
	Set           string // SetExternalID extracted from gallery
	Lang          items.LangCode
	Rarity        string // Full name from title attribute (e.g., "Prismatic Secret Rare")
	RarityCode    string // Short code from visible text (e.g., "PScR")
	Edition       string // "1st Edition" or "Unlimited Edition"
	PrintURLSmall string // Thumbnail URL (~120px)
	PrintURLLarge string // Original full-size URL (derived from thumbnail)
}

// parseGallery parses a Card Gallery page and returns gallery entries.
// It excludes proxy images and broken images (no <a class="image">).
func parseGallery(e *colly.HTMLElement) []galleryEntry {
	galleryIDToLang := map[string]items.LangCode{
		"card-gallery--EN": items.EN,
		"card-gallery--NA": items.EN,
		"card-gallery--EU": items.EN,
		"card-gallery--OC": items.EN,
		"card-gallery--FR": items.FR,
		"card-gallery--FC": items.FR,
		"card-gallery--DE": items.DE,
		"card-gallery--IT": items.IT,
		"card-gallery--PT": items.PT,
		"card-gallery--SP": items.SP,
		"card-gallery--JP": items.JP,
		"card-gallery--JA": items.JP,
		"card-gallery--AE": items.AE,
		"card-gallery--KR": items.KR,
		"card-gallery--TC": items.TC,
		"card-gallery--SC": items.SC,
	}

	var entries []galleryEntry

	e.ForEach("div.card-gallery", func(_ int, galleryDiv *colly.HTMLElement) {
		galleryID, exists := galleryDiv.DOM.Attr("id")
		if !exists {
			return
		}
		lang, ok := galleryIDToLang[galleryID]
		if !ok {
			return
		}

		galleryDiv.ForEach("li.gallerybox", func(_ int, li *colly.HTMLElement) {
			// Check for proxy (skip if Official_Proxy link exists)
			proxyLink := li.DOM.Find("div.gallerytext a[href*='Official_Proxy']")
			if proxyLink.Length() > 0 {
				return
			}

			// Check for broken image (skip if no <a class="image"> in thumb)
			imageLink := li.DOM.Find("div.thumb a.image")
			if imageLink.Length() == 0 {
				return
			}

			// Extract thumbnail URL
			img := imageLink.Find("img")
			if img.Length() == 0 {
				return
			}
			thumbURL, exists := img.Attr("src")
			if !exists || thumbURL == "" {
				return
			}

			// Extract code (first <a> in gallerytext)
			galleryText := li.DOM.Find("div.gallerytext")
			if galleryText.Length() == 0 {
				return
			}
			codeLink := galleryText.Find("a").First()
			code := strings.TrimSpace(codeLink.Text())
			if code == "" {
				return
			}

			// Extract rarity (title attribute from rarity link)
			// Rarity links are typically in parentheses after the code and have a title attribute
			var rarity, rarityCode string
			galleryText.Find("a[title]").Each(func(_ int, a *goquery.Selection) {
				title, _ := a.Attr("title")
				// Skip edition links and set links
				if strings.Contains(title, "Edition") || strings.Contains(title, "Pack") || strings.Contains(title, "Box") {
					return
				}
				// Rarity titles typically contain "Rare", "Common", or are specific rarity names
				if strings.Contains(title, "Rare") || strings.Contains(title, "Common") ||
					strings.Contains(title, "Secret") || strings.Contains(title, "Ultimate") ||
					strings.Contains(title, "Collector") || strings.Contains(title, "Parallel") {
					if rarity == "" {
						rarity = title
						rarityCode = strings.TrimSpace(a.Text())
					}
				}
			})

			// Extract set from <i><a href="/wiki/Set_Name">...</a></i>
			// The set link is typically wrapped in <i> tags in the gallery text
			var set string
			galleryText.Find("i a").Each(func(_ int, a *goquery.Selection) {
				href, _ := a.Attr("href")
				if href != "" && strings.HasPrefix(href, "/wiki/") {
					set = wikiPageName(href)
				}
			})

			// Fallback: try any remaining link that isn't code, rarity, or edition
			if set == "" {
				galleryText.Find("a").Each(func(_ int, a *goquery.Selection) {
					href, _ := a.Attr("href")
					title, _ := a.Attr("title")
					if strings.Contains(href, "1st_Edition") || strings.Contains(href, "Unlimited_Edition") {
						return
					}
					if strings.Contains(title, "Edition") || strings.Contains(title, "Rare") ||
						strings.Contains(title, "Common") || strings.Contains(title, "Secret") ||
						strings.Contains(title, "Ultimate") || strings.Contains(title, "Collector") ||
						strings.Contains(title, "Parallel") {
						return
					}
					linkText := strings.TrimSpace(a.Text())
					if linkText == code || linkText == rarityCode {
						return
					}
					if set == "" {
						set = wikiPageName(href)
					}
				})
			}

			// Extract edition
			var edition string
			galleryText.Find("a").Each(func(_ int, a *goquery.Selection) {
				href, _ := a.Attr("href")
				if strings.Contains(href, "1st_Edition") {
					edition = "1st Edition"
				} else if strings.Contains(href, "Unlimited_Edition") {
					edition = "Unlimited Edition"
				}
			})

			// Derive original URL from thumbnail
			originalURL := deriveOriginalURL(thumbURL)

			entries = append(entries, galleryEntry{
				Code:          code,
				Set:           set,
				Lang:          lang,
				Rarity:        rarity,
				RarityCode:    rarityCode,
				Edition:       edition,
				PrintURLSmall: thumbURL,
				PrintURLLarge: originalURL,
			})
		})
	})

	return entries
}

// deriveOriginalURL converts a MediaWiki thumbnail URL to the original full-size URL.
// Example:
//
//	Input:  https://ms.yugipedia.com//thumb/f/f6/Foo.png/120px-Foo.png
//	Output: https://ms.yugipedia.com//f/f6/Foo.png
func wikiPageName(href string) string {
	// Extract page name from href: /wiki/Page_Name → Page Name
	prefix := "/wiki/"
	if !strings.HasPrefix(href, prefix) {
		return ""
	}
	name := href[len(prefix):]
	name = strings.ReplaceAll(name, "_", " ")
	name, _, _ = strings.Cut(name, "?")
	name, _, _ = strings.Cut(name, "#")
	return name
}

func parseGalleryLink(e *colly.HTMLElement) string {
	var galleryURL string
	e.ForEach("div.hlist ul li a", func(_ int, a *colly.HTMLElement) {
		if strings.TrimSpace(a.Text) == "Gallery" {
			href, _ := a.DOM.Attr("href")
			if strings.Contains(href, "Card_Gallery:") {
				galleryURL = href
			}
		}
	})
	return galleryURL
}

func deriveOriginalURL(thumbnailURL string) string {
	// Remove /thumb from path
	idx := strings.Index(thumbnailURL, "/thumb/")
	if idx == -1 {
		return thumbnailURL
	}
	withoutThumb := thumbnailURL[:idx] + thumbnailURL[idx+len("/thumb"):]

	// Remove the last segment /{width}px-{filename}
	lastSlash := strings.LastIndex(withoutThumb, "/")
	if lastSlash == -1 {
		return withoutThumb
	}
	return withoutThumb[:lastSlash]
}
