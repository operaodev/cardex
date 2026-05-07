package providers

import (
	"regexp"
	"strings"

	"github.com/operaodev/cardex/internal/cards"
	"github.com/operaodev/cardex/internal/search"
)


// wikitextNameKeys mapea las claves de nombre en wikitext al LangCode interno.
var wikitextNameKeys = map[string]cards.LangCode{
	"fr_name": cards.FR,
	"de_name": cards.DE,
	"it_name": cards.IT,
	"pt_name": cards.PT,
	"es_name": cards.SP,
	"ja_name": cards.JP,
	"ko_name": cards.KR,
	"tc_name": cards.TC,
	"sc_name": cards.SC,
}

// wikitextDescKeys mapea las claves de descripción en wikitext al LangCode interno.
var wikitextDescKeys = map[string]cards.LangCode{
	"text":    cards.EN,
	"fr_text": cards.FR,
	"de_text": cards.DE,
	"it_text": cards.IT,
	"pt_text": cards.PT,
	"es_text": cards.SP,
	"ja_text": cards.JP,
	"ko_text": cards.KR,
	"tc_text": cards.TC,
	"sc_text": cards.SC,
}

// wikitextSetKeys mapea las claves de sets al LangCode.
// Todos los sets ingleses (en/na/eu/au) se consolidan bajo EN.
var wikitextSetKeys = map[string]cards.LangCode{
	"en_sets": cards.EN,
	"na_sets": cards.EN,
	"eu_sets": cards.EN,
	"au_sets": cards.EN,
	"fr_sets": cards.FR,
	"de_sets": cards.DE,
	"it_sets": cards.IT,
	"pt_sets": cards.PT,
	"sp_sets": cards.SP,
	"jp_sets": cards.JP,
	"kr_sets": cards.KR,
	"sc_sets": cards.SC,
}

// extractAllFields parsea el wikitext y retorna un map[key]value con todos los
// campos encontrados. Soporta valores multi-línea.
//
// Funciona dividiendo el wikitext por "\n| " (separador de campos en la
// plantilla CardTable2) y parseando cada fragmento como key = value.
func extractAllFields(wikitext string) map[string]string {
	fields := make(map[string]string)

	// Dividir por "\n| " para obtener cada bloque de campo.
	// Agregamos \n al inicio para que el primer campo también matchee.
	chunks := strings.Split("\n"+wikitext, "\n| ")

	for _, chunk := range chunks[1:] { // saltar el preámbulo ({{CardTable2)
		before, after, ok := strings.Cut(chunk, "=")
		if !ok {
			continue
		}

		key := strings.TrimSpace(before)
		if !isValidFieldKey(key) {
			continue
		}

		value := strings.TrimSpace(after)

		// Limpiar el cierre de template del último campo
		if idx := strings.Index(value, "\n}}"); idx >= 0 {
			value = strings.TrimSpace(value[:idx])
		}

		value = cleanWikiLinks(value)
		value = cleanRubyMarkup(value)

		if value != "" {
			fields[key] = value
		}
	}

	return fields
}

// isValidFieldKey verifica que la clave solo contenga letras minúsculas,
// dígitos, underscores o barras (para claves como "m/s/t").
func isValidFieldKey(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '/') {
			return false
		}
	}
	return true
}

// extractNames extrae los nombres localizados desde los campos parseados.
func extractNames(fields map[string]string) map[cards.LangCode]string {
	names := make(map[cards.LangCode]string)

	for key, lang := range wikitextNameKeys {
		if v, ok := fields[key]; ok {
			names[lang] = v
		}
	}

	return names
}

// extractDescriptions extrae las descripciones localizadas desde los campos parseados.
func extractDescriptions(fields map[string]string) map[cards.LangCode]string {
	descriptions := make(map[cards.LangCode]string)

	for key, lang := range wikitextDescKeys {
		if v, ok := fields[key]; ok {
			descriptions[lang] = v
		}
	}

	return descriptions
}

// extractPrintedCards extrae los sets/prints por idioma y los convierte
// a []search.ResultPrintedCard.
func extractPrintedCards(fields map[string]string) []search.ResultPrintedCard {
	var printed []search.ResultPrintedCard

	for key, lang := range wikitextSetKeys {
		setsText, ok := fields[key]
		if !ok {
			continue
		}

		entries := parseSetEntries(setsText, lang)
		printed = append(printed, entries...)
	}

	return printed
}

// parseSetEntries parsea líneas con formato: CÓDIGO; Nombre del Set; Rareza1, Rareza2
// Cada rareza genera un ResultPrintedCard separado.
func parseSetEntries(setsText string, lang cards.LangCode) []search.ResultPrintedCard {
	var entries []search.ResultPrintedCard

	lines := strings.Split(setsText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ";", 3)
		if len(parts) < 3 {
			continue
		}

		code := strings.TrimSpace(parts[0])
		setName := strings.TrimSpace(parts[1])
		raritiesStr := strings.TrimSpace(parts[2])

		// Cada rareza genera una entrada separada
		rarities := strings.Split(raritiesStr, ",")
		for _, r := range rarities {
			r = strings.TrimSpace(r)
			if r == "" {
				continue
			}

			entries = append(entries, search.ResultPrintedCard{
				Code:    code,
				SetName: setName,
				Rarity:  cards.Rarity(r),
				Lang:    lang,
				TCG:     cards.TCGYugioh,
			})
		}
	}

	return entries
}

// cleanWikiLinks limpia los enlaces wiki del texto.
// [[enlace|texto]] -> texto, [[enlace]] -> enlace
var wikiLinkWithPipe = regexp.MustCompile(`\[\[[^\|\]]+\|([^\]]+)\]\]`)
var wikiLinkSimple = regexp.MustCompile(`\[\[([^\]]+)\]\]`)

func cleanWikiLinks(text string) string {
	text = wikiLinkWithPipe.ReplaceAllString(text, "$1")
	text = wikiLinkSimple.ReplaceAllString(text, "$1")
	return text
}

// cleanRubyMarkup limpia el markup Ruby de MediaWiki (furigana japonés).
// {{Ruby|剣|つるぎ}} -> 剣
var rubyPattern = regexp.MustCompile(`\{\{Ruby\|([^|]+)\|[^}]+\}\}`)

func cleanRubyMarkup(text string) string {
	return rubyPattern.ReplaceAllString(text, "$1")
}

// parseWikitext toma el wikitext crudo de Yugipedia y retorna un ResultCard
// con toda la información extraída (nombres, descripciones, sets).
func parseWikitext(wikitext string) search.ResultCard {
	fields := extractAllFields(wikitext)

	return search.ResultCard{
		Names:        extractNames(fields),
		Descriptions: extractDescriptions(fields),
		PrintedCards: extractPrintedCards(fields),
	}
}
