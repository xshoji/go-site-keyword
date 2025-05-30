package analyzer

import (
	"testing"

	"github.com/xshoji/go-site-keyword/internal/parser"
	"github.com/xshoji/go-site-keyword/pkg/config"
)

type dummyNormalizer struct{}

func (d dummyNormalizer) NormalizeKeyword(word string) string { return word }

// テスト用: HTML文字列からAnalyzerを生成
func NewAnalyzerFromHTML(html string, cfg config.Config) *Analyzer {
	doc, err := parser.ParseHTMLDocument(html)
	if err != nil {
		panic(err)
	}
	return &Analyzer{
		URL:          "dummy",
		responseBody: []byte(html),
		doc:          doc,
		Config:       cfg,
	}
}

// goquery.Documentのラッパーをテスト用に生成
func ParseHTMLDocumentForTest(html string) (*parser.HTMLDocument, error) {
	return parser.ParseHTMLDocument(html)
}

func TestAnalyzer_FetchTitleAndMeta(t *testing.T) {
	html := `<html><head><title>TestTitle</title><meta name="description" content="desc"><meta name="keywords" content="go, test"></head><body><h1>見出し</h1></body></html>`
	doc := NewAnalyzerFromHTML(html, config.DefaultConfig())
	title, _ := doc.FetchTitle()
	if title != "TestTitle" {
		t.Errorf("expected 'TestTitle', got '%s'", title)
	}
	meta, _ := doc.FetchMetaTags()
	if meta["description"] != "desc" {
		t.Errorf("expected 'desc', got '%s'", meta["description"])
	}
	if meta["keywords"] != "go, test" {
		t.Errorf("expected 'go, test', got '%s'", meta["keywords"])
	}
}

func TestAnalyzer_GetTopKeywords_English(t *testing.T) {
	html := `<html><head><title>Go Test</title><meta name="keywords" content="go, test, code"></head><body><h1>Go Test</h1></body></html>`
	doc := NewAnalyzerFromHTML(html, config.DefaultConfig())
	stopWords := map[string]int{"the": 0}
	keywords, err := doc.GetTopKeywords(3, stopWords, func(s string) string { return s })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keywords) == 0 {
		t.Error("expected keywords, got none")
	}
}

func TestAnalyzer_GetTopKeywords_Japanese(t *testing.T) {
	html := `<html><head><title>日本語 テスト</title></head><body><h1>日本語 テスト</h1></body></html>`
	doc := NewAnalyzerFromHTML(html, config.DefaultConfig())
	keywords, err := doc.GetTopKeywords(3, map[string]int{}, func(s string) string { return s })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keywords) == 0 {
		t.Error("expected keywords, got none")
	}
}
