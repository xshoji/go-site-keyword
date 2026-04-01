package analyzer

import (
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/xshoji/go-site-keyword/internal/fetcher"
	"github.com/xshoji/go-site-keyword/internal/language"
	"github.com/xshoji/go-site-keyword/internal/language/english"
	"github.com/xshoji/go-site-keyword/internal/language/japanese"
	"github.com/xshoji/go-site-keyword/internal/parser"
	"github.com/xshoji/go-site-keyword/internal/scoring"
	"github.com/xshoji/go-site-keyword/pkg/config"
	"github.com/xshoji/go-site-keyword/pkg/types"
)

type PageData struct {
	Title       string
	MetaTags    map[string]string
	MainContent string
}

type Analyzer struct {
	URL          string
	responseBody []byte
	doc          *parser.HTMLDocument
	Config       config.Config
}

func NewAnalyzer(url string, cfg config.Config) (*Analyzer, error) {
	res, err := fetcher.FetchURL(url, int(cfg.Timeout.Seconds()))
	if err != nil {
		return nil, err
	}
	doc, err := parser.ParseHTMLDocument(string(res.Body))
	if err != nil {
		return nil, err
	}
	return &Analyzer{
		URL:          res.URL,
		responseBody: res.Body,
		doc:          doc,
		Config:       cfg,
	}, nil
}

func (a *Analyzer) FetchTitle() (string, error) {
	titles := a.doc.FetchTags("title")
	if len(titles) == 0 {
		return "", nil
	}
	return titles[0], nil
}

func (a *Analyzer) FetchMetaTags() (map[string]string, error) {
	return a.doc.FetchMetaTags(), nil
}

func (a *Analyzer) FetchMainContent() (string, error) {
	segments := a.doc.ExtractStructuredText()
	var parts []string
	for _, seg := range segments {
		if seg.Source == "body" || seg.Source == "h1" || seg.Source == "h2" || seg.Source == "h3" {
			parts = append(parts, seg.Text)
		}
	}
	return strings.Join(parts, " "), nil
}

func (a *Analyzer) CollectPageData() (*PageData, error) {
	title, _ := a.FetchTitle()
	meta := a.doc.FetchMetaTags()
	content, _ := a.FetchMainContent()
	return &PageData{
		Title:       title,
		MetaTags:    meta,
		MainContent: content,
	}, nil
}

// sourceWeight returns the config weight for a given TextSegment source.
func (a *Analyzer) sourceWeight(source string) int {
	w := a.Config.ScoreWeights
	switch source {
	case "title":
		return w.Title
	case "meta_keywords":
		return w.MetaKeyword
	case "meta_description":
		return w.MetaDescription
	case "h1":
		return w.H1
	case "h2", "h3":
		return w.H2H3
	case "emphasis":
		return w.Emphasis
	case "anchor":
		return w.Anchor
	case "alt":
		return w.Alt
	case "body":
		return w.Body
	default:
		return 1
	}
}

// GetTopKeywords extracts keywords using the new multi-layer pipeline.
// stopWords and normalizeKeyword are used for English text normalization.
func (a *Analyzer) GetTopKeywords(n int, stopWords map[string]int, normalizeKeyword func(string) string) ([]scoring.KeywordWithScore, error) {
	cfg := a.Config
	if n <= 0 {
		n = cfg.MaxKeywords
	}

	segments := a.doc.ExtractStructuredText()
	totalSegments := len(segments)

	// Accumulate scores: normalized keyword → raw score (float64)
	scoreMap := make(map[string]float64)
	// Track original surface form
	originalMap := make(map[string]string)
	// Track first occurrence position (0.0 - 1.0)
	firstPosMap := make(map[string]float64)

	for i, seg := range segments {
		position := 0.5
		if totalSegments > 1 {
			position = float64(i) / float64(totalSegments-1)
		}
		weight := a.sourceWeight(seg.Source)

		// Extract keywords with frequency from the segment text
		var freqMap map[string]int
		if language.ContainsJapanese(seg.Text) {
			freqMap = japanese.ExtractJapaneseKeywordsWithFrequency(seg.Text)
		} else {
			freqMap = english.ExtractEnglishKeywordsWithFrequency(seg.Text, stopWords, normalizeKeyword)
		}

		for kw, freq := range freqMap {
			scoreMap[kw] += float64(freq) * float64(weight)
			// Track original form (prefer longer surface)
			if existing, ok := originalMap[kw]; !ok || len(kw) > len(existing) {
				originalMap[kw] = kw
			}
			// Track first position
			if _, ok := firstPosMap[kw]; !ok {
				firstPosMap[kw] = position
			}
		}
	}

	return scoring.RankKeywordsAdvanced(scoreMap, originalMap, firstPosMap, n, cfg.MinScore), nil
}

// extractKeywords: 言語自動判定して適切な抽出関数を呼ぶ
func extractKeywords(text string, stopWords map[string]int, normalizeKeyword func(string) string) []string {
	if language.ContainsJapanese(text) {
		return japanese.ExtractJapaneseKeywords(text)
	}
	return english.ExtractEnglishKeywords(text, stopWords, normalizeKeyword)
}

// ページ取得の分離
func FetchPage(url string, timeout time.Duration) (*http.Response, error) {
	client := &http.Client{Timeout: timeout}
	return client.Get(url)
}

// 文書解析の分離
func ParseDocument(body []byte) (*goquery.Document, error) {
	return goquery.NewDocumentFromReader(strings.NewReader(string(body)))
}

// キーワード抽出の分離
func ExtractKeywords(content string, isJapanese bool, stopWords map[string]int, normalizeKeyword func(string) string) ([]string, error) {
	if isJapanese || language.ContainsJapanese(content) {
		return japanese.ExtractJapaneseKeywords(content), nil
	}
	return english.ExtractEnglishKeywords(content, stopWords, normalizeKeyword), nil
}

// ExtractKeywordsWithFrequency テキストからキーワードとその頻度を抽出します
func ExtractKeywordsWithFrequency(text string, stopWords map[string]int, normalizeKeyword func(string) string) []scoring.KeywordWithScore {
	words := strings.Fields(strings.ToLower(text))
	wordFreq := map[string]int{}
	normalizedWords := map[string][]string{}
	normalizedScores := map[string]int{}
	for _, w := range words {
		if _, skip := stopWords[w]; skip || len(w) <= 1 || w == "-" {
			continue
		}
		norm := normalizeKeyword(w)
		if _, skip := stopWords[norm]; !skip && len(norm) > 1 && norm != "-" {
			wordFreq[w]++
			if norm != w {
				normalizedWords[norm] = append(normalizedWords[norm], w)
			}
			normalizedScores[norm]++
		}
	}
	var result []scoring.KeywordWithScore
	for norm, score := range normalizedScores {
		bestWord := norm
		bestScore := 0
		if originals, exists := normalizedWords[norm]; exists && len(originals) > 0 {
			for _, original := range originals {
				if wordFreq[original] > bestScore {
					bestWord = original
					bestScore = wordFreq[original]
				}
			}
		}
		result = append(result, scoring.KeywordWithScore{
			Keyword: bestWord,
			Score:   score,
		})
	}
	return result
}

// AnalyzerのメソッドとしてConfigのストップワード・正規化辞書を使う例
func (a *Analyzer) GetTopKeywordsWithDefaultConfig(n int) ([]scoring.KeywordWithScore, error) {
	cfg := a.Config
	stopWords := cfg.EnglishStopWords
	pluralSingularMap := cfg.PluralSingularMap
	invariantWords := cfg.InvariantWords
	normalize := func(word string) string {
		return english.NormalizeEnglishKeyword(word, pluralSingularMap, invariantWords)
	}
	mainContent, _ := a.FetchMainContent()
	return ExtractKeywordsWithFrequency(mainContent, stopWords, normalize), nil
}

// stopWords, normalizeKeyword をConfigから自動で利用するバージョン
func (a *Analyzer) GetTopKeywordsAuto(n int) ([]scoring.KeywordWithScore, error) {
	cfg := a.Config
	stopWords := cfg.EnglishStopWords
	pluralSingularMap := cfg.PluralSingularMap
	invariantWords := cfg.InvariantWords
	normalize := func(word string) string {
		return english.NormalizeEnglishKeyword(word, pluralSingularMap, invariantWords)
	}
	return a.GetTopKeywords(n, stopWords, normalize)
}

// GetAnalysisResult はウェブページの解析結果を返します
func (a *Analyzer) GetAnalysisResult(maxKeywords int) (*types.AnalysisResult, error) {
	result := &types.AnalysisResult{}
	var lastErr error

	// タイトルを取得
	title, err := a.FetchTitle()
	if err != nil {
		lastErr = err
	} else if title != "" {
		result.Title = title
	}

	// メタタグを取得
	meta, err := a.FetchMetaTags()
	if err != nil {
		lastErr = err
	} else if len(meta) > 0 {
		result.MetaTags = meta
	}

	// キーワードを取得
	keywordsWithScores, err := a.GetTopKeywordsAuto(maxKeywords)
	if err != nil {
		lastErr = err
	} else if len(keywordsWithScores) > 0 {
		result.Keywords = convertToTypeKeywords(keywordsWithScores)
	}

	// 何かしらのデータが取得できていれば結果を返す
	if result.Title != "" || len(result.MetaTags) > 0 || len(result.Keywords) > 0 {
		return result, lastErr
	}

	// 何もデータが取得できなかった場合はエラーを返す
	if lastErr != nil {
		return nil, lastErr
	}

	return result, nil
}

// convertToTypeKeywords は scoring.KeywordWithScore を types.KeywordWithScore に変換します
func convertToTypeKeywords(input []scoring.KeywordWithScore) []types.KeywordWithScore {
	var result []types.KeywordWithScore
	for _, kws := range input {
		result = append(result, types.KeywordWithScore{
			Keyword: kws.Keyword,
			Score:   kws.Score,
		})
	}
	return result
}
