package japanese

import (
	"strings"
	"unicode"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

var nounSubcategories = map[string]bool{
	"一般":     true,
	"固有名詞":   true,
	"サ変接続":   true,
	"形容動詞語幹": true,
	"接尾":     true,
}

// Japanese stop words: common words with little keyword value
var japaneseStopWords = map[string]bool{
	"さん": true, "ちゃん": true, "くん": true, "様": true, "殿": true,
	"者": true, "方": true, "側": true, "上": true, "下": true, "中": true, "内": true, "外": true,
	"的": true, "用": true, "別": true, "化": true, "性": true, "感": true,
	"等": true, "他": true, "系": true, "間": true, "分": true, "数": true, "回": true,
	"今": true, "前": true, "後": true, "次": true, "初": true, "末": true,
	"ため": true, "こと": true, "もの": true, "それ": true, "これ": true, "あれ": true,
	"ここ": true, "そこ": true, "あそこ": true, "どこ": true,
	"よう": true, "ところ": true, "わけ": true, "はず": true,
	"つもり": true, "まま": true, "ほう": true, "ほか": true,
	"月": true, "日": true, "年": true, "時": true, "火": true, "水": true, "木": true, "金": true, "土": true,
	"つ": true, "ん": true, "の": true, "に": true,
}

// isJapaneseStopWord checks if a surface form is a Japanese stop word
func isJapaneseStopWord(surface string) bool {
	return japaneseStopWords[surface]
}

func isNounToken(features []string) bool {
	if len(features) == 0 || features[0] != "名詞" {
		return false
	}
	if len(features) <= 1 {
		return false
	}
	return nounSubcategories[features[1]]
}

func shouldKeepSurface(surface string) bool {
	runes := []rune(surface)
	if len(runes) == 1 && !unicode.In(runes[0], unicode.Han) {
		return false
	}
	if isSymbolOrPunctuation(surface) {
		return false
	}
	if containsSymbol(surface) {
		return false
	}
	if isJapaneseStopWord(surface) {
		return false
	}
	return true
}

func ExtractJapaneseKeywords(text string) []string {
	freq := ExtractJapaneseKeywordsWithFrequency(text)
	result := make([]string, 0, len(freq))
	for k := range freq {
		result = append(result, k)
	}
	return result
}

func ExtractJapaneseKeywordsWithFrequency(text string) map[string]int {
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		return map[string]int{}
	}
	tokens := t.Tokenize(text)

	// Build runs of consecutive noun tokens
	type nounToken struct {
		surface  string
		features []string
	}
	var runs [][]nounToken
	var currentRun []nounToken

	for _, token := range tokens {
		features := token.Features()
		if isNounToken(features) {
			currentRun = append(currentRun, nounToken{surface: token.Surface, features: features})
		} else {
			if len(currentRun) > 0 {
				runs = append(runs, currentRun)
				currentRun = nil
			}
		}
	}
	if len(currentRun) > 0 {
		runs = append(runs, currentRun)
	}

	freq := make(map[string]int)
	normalizedMap := make(map[string]string)

	addKeyword := func(surface string) {
		if !shouldKeepSurface(surface) {
			return
		}
		normalized := strings.ToLower(surface)
		freq[normalized]++
		if existing, ok := normalizedMap[normalized]; !ok || len(surface) > len(existing) {
			normalizedMap[normalized] = surface
		}
	}

	for _, run := range runs {
		// Add individual nouns (skip suffix-only tokens like 者, さん)
		for _, nt := range run {
			if nt.features[1] != "接尾" {
				addKeyword(nt.surface)
			}
		}
		// Add compound nouns for runs of length >= 2
		if len(run) >= 2 {
			// Full compound
			var compound strings.Builder
			for _, nt := range run {
				compound.WriteString(nt.surface)
			}
			addKeyword(compound.String())

			// Also add sliding window compounds of length 2 and 3
			// to capture sub-compounds like "人工知能" from longer runs.
			// Skip windows that start with a suffix token or end just
			// before a suffix token (e.g., "廣川容疑" when "者" follows).
			for winLen := 2; winLen <= 3 && winLen < len(run); winLen++ {
				for start := 0; start+winLen <= len(run); start++ {
					if run[start].features[1] == "接尾" {
						continue
					}
					// Skip if the next token after this window is a suffix
					// (the window would be cutting a natural compound short)
					endIdx := start + winLen
					if endIdx < len(run) && run[endIdx].features[1] == "接尾" {
						continue
					}
					var sub strings.Builder
					for j := start; j < start+winLen; j++ {
						sub.WriteString(run[j].surface)
					}
					addKeyword(sub.String())
				}
			}
		}
	}

	// Build result with original casing
	result := make(map[string]int, len(freq))
	for norm, count := range freq {
		if original, ok := normalizedMap[norm]; ok {
			result[original] = count
		} else {
			result[norm] = count
		}
	}
	return result
}

func isSymbolOrPunctuation(text string) bool {
	if text == "" {
		return false
	}
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) ||
			unicode.In(r, unicode.Hiragana, unicode.Katakana, unicode.Han) {
			return false
		}
	}
	return true
}

// containsSymbol returns true if the text contains any punctuation or bracket characters.
// This filters out tokens like "(木", "ページ)" that the tokenizer produces.
func containsSymbol(text string) bool {
	for _, r := range text {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) ||
			r == '(' || r == ')' || r == '（' || r == '）' ||
			r == '「' || r == '」' || r == '『' || r == '』' {
			return true
		}
	}
	return false
}
