package scoring

import (
	"sort"
)

type KeywordWithScore struct {
	Keyword string
	Score   int
}

// RankKeywordsByScore はキーワードをスコア順にランク付けします
func RankKeywordsByScore(scoreMap map[string]int, originalMap map[string]string, limit int) []KeywordWithScore {
	type kv struct {
		Key   string
		Value int
	}
	var sorted []kv
	for k, v := range scoreMap {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	var result []KeywordWithScore
	for _, kv := range sorted {
		originalKey := kv.Key
		if original, ok := originalMap[kv.Key]; ok {
			originalKey = original
		}
		result = append(result, KeywordWithScore{Keyword: originalKey, Score: kv.Value})
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result
}

// KeywordScore tracks accumulated score for a keyword
type KeywordScore struct {
	Keyword     string
	RawScore    float64
	Occurrences int
	FirstPos    float64 // 0.0 = start of document, 1.0 = end
}

// CalculatePositionBoost returns a boost factor based on position in document.
// Earlier positions get higher boost: first third → 1.5, middle → 1.2, last → 1.0
func CalculatePositionBoost(position float64) float64 {
	if position < 0.33 {
		return 1.5
	}
	if position < 0.66 {
		return 1.2
	}
	return 1.0
}

// CalculateLengthPenalty returns a penalty factor based on keyword character length.
// 1 char → 0.3, 2 chars → 0.7, 3+ chars → 1.0
// For Japanese text (rune count is used).
func CalculateLengthPenalty(keyword string) float64 {
	runeLen := len([]rune(keyword))
	switch {
	case runeLen <= 1:
		return 0.3
	case runeLen == 2:
		return 0.7
	default:
		return 1.0
	}
}

// RankKeywordsAdvanced ranks keywords using position boost and length penalty.
// scoreMap: keyword → accumulated raw score
// originalMap: normalized keyword → original surface form
// firstPosMap: keyword → first occurrence position (0.0-1.0)
// limit: max results (0 = unlimited)
// minScore: minimum score threshold (keywords below this are excluded)
func RankKeywordsAdvanced(scoreMap map[string]float64, originalMap map[string]string, firstPosMap map[string]float64, limit int, minScore float64) []KeywordWithScore {
	type scored struct {
		key   string
		score float64
	}
	var items []scored
	for k, rawScore := range scoreMap {
		pos := 0.5
		if p, ok := firstPosMap[k]; ok {
			pos = p
		}
		finalScore := rawScore * CalculatePositionBoost(pos) * CalculateLengthPenalty(k)
		if finalScore < minScore {
			continue
		}
		items = append(items, scored{k, finalScore})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].score > items[j].score
	})

	var result []KeywordWithScore
	for _, item := range items {
		keyword := item.key
		if orig, ok := originalMap[item.key]; ok {
			keyword = orig
		}
		result = append(result, KeywordWithScore{
			Keyword: keyword,
			Score:   int(item.score + 0.5), // round to int
		})
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result
}
