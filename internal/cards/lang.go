package cards

import "fmt"

type LangCode string
type LangName string

const (
	EN LangCode = "en" // English
	FR LangCode = "fr" // French
	DE LangCode = "de" // German
	IT LangCode = "it" // Italian
	PT LangCode = "pt" // Portuguese
	SP LangCode = "sp" // Spanish

	JP LangCode = "jp" // Japanese
	AE LangCode = "ae" // Asian-English
	KR LangCode = "kr" // Korean
	TC LangCode = "tc" // Traditional Chinese
	SC LangCode = "sc" // Simplified Chinese
)

const (
	English    LangName = "English"
	French     LangName = "Français"
	German     LangName = "Deutsch"
	Italian    LangName = "Italiano"
	Portuguese LangName = "Português"
	Spanish    LangName = "Español"

	Japanese           LangName = "日本語"
	AsianEnglish       LangName = "English (Asia)"
	Korean             LangName = "한국어"
	TraditionalChinese LangName = "繁體中文"
	SimplifiedChinese  LangName = "简体中文"
)

func GetLangName(code LangCode) (LangName, error) {
	switch code {
	case EN:
		return English, nil
	case FR:
		return French, nil
	case DE:
		return German, nil
	case IT:
		return Italian, nil
	case PT:
		return Portuguese, nil
	case SP:
		return Spanish, nil
	case JP:
		return Japanese, nil
	case AE:
		return AsianEnglish, nil
	case KR:
		return Korean, nil
	case TC:
		return TraditionalChinese, nil
	case SC:
		return SimplifiedChinese, nil
	default:
		return "", fmt.Errorf("unsupported language code: %s", code)
	}
}