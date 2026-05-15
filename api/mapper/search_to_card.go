package mapper

import (
	"github.com/operaodev/cardex/internal/cards"
	"github.com/operaodev/cardex/internal/search"
)

// ResultToCards convierte un ResultCard en múltiples cards.Card.
//
// Lógica de mapeo:
//   - Solo se generan Cards para resultados con impresiones físicas (PrintedCards).
//   - Se genera UNA Card por cada PrintedCard del resultado.
//   - El idioma de cada PrintedCard determina qué nombre y descripción usar.
//   - Si no existe traducción para ese idioma, se usa el inglés como fallback.
//   - La identidad única es: ExternalID + Code + Lang + Rarity (idx_card_identity).
func ResultToCards(result search.ResultCard) []cards.Card {
	englishName := result.Names[cards.EN]

	out := make([]cards.Card, 0, len(result.PrintedCards))

	for _, printed := range result.PrintedCards {
		lang := printed.Lang

		// Resolver nombre: usar el idioma del print, fallback a inglés
		name := result.Names[lang]
		if name == "" {
			name = englishName
		}

		// Resolver descripción: usar el idioma del print, fallback a inglés
		description := result.Descriptions[lang]
		if description == "" {
			description = result.Descriptions[cards.EN]
		}

		// TCG viene del PrintedCard (más específico), fallback al del ResultCard
		tcg := printed.TCG
		if tcg == "" {
			tcg = result.TCG
		}

		out = append(out, cards.Card{
			TCG:         tcg,
			ExternalID:  result.ExternalID,
			EnglishName: englishName,
			Name:        name,
			Description: description,
			Lang:        lang,
			Type:        result.Type,
			Subtypes:    result.Subtypes,
			Archetype:   result.Archetype,
			Sources:     result.Sources,
			CardImages:  result.Images,
			// Datos de la impresión específica
			Code:    printed.Code,
			SetName: printed.SetName,
			Rarity:  printed.Rarity,
		})
	}

	return out
}

// ResultsToCards convierte un slice de ResultCard en un slice plano de cards.Card.
// Útil para el flujo de importación masiva (FetchAllCards).
func ResultsToCards(results []search.ResultCard) []cards.Card {
	out := make([]cards.Card, 0, len(results))
	for _, result := range results {
		out = append(out, ResultToCards(result)...)
	}
	return out
}
