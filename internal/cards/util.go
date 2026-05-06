package cards

import "fmt"


// GenerateCardUniqueID("ygo", "en", "12345") // "ygo-en-12345"
func GenerateCardUniqueID(tcg TCGType, lang LangCode, externalID string) string {
	return fmt.Sprintf("%s-%s-%s", tcg, lang, externalID)
}

// GenerateCardSharedID("ygo", "12345") // "ygo-12345"
func GenerateCardSharedID(tcg TCGType, externalID string) string {
	return fmt.Sprintf("%s-%s", tcg, externalID)
}

// GeneratePrintUniqueID("ygo", "en", "RAO5-EN001", "common") // "ygo-en-RAO5-EN001-common"
func GeneratePrintUniqueID(tcg TCGType, lang LangCode, externalCode string, rarity string) string {
	return fmt.Sprintf("%s-%s-%s-%s", tcg, lang, externalCode, rarity)
}

// GeneratePrintSharedID("ygo", "RAO5", "001", "common") // "ygo-RAO5-001-common"
func GeneratePrintSharedID(tcg TCGType, setCode, setNumber, rarity string) string {
	return fmt.Sprintf("%s-%s-%s-%s", tcg, setCode, setNumber, rarity)
}
