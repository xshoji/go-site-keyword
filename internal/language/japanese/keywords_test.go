package japanese

import (
	"testing"
)

func TestExtractJapaneseKeywords(t *testing.T) {
	text := "これはテスト用の文章です。Go言語と形態素解析を使います。"
	keywords := ExtractJapaneseKeywords(text)
	if len(keywords) == 0 {
		t.Error("expected some keywords, got none")
	}
	found := false
	for _, k := range keywords {
		if k == "言語" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected '言語' in keywords, got %v", keywords)
	}
}

func TestExtractJapaneseKeywords_CompoundNoun(t *testing.T) {
	text := "人工知能と機械学習は現代の技術です。自然言語処理も重要な分野です。"
	keywords := ExtractJapaneseKeywords(text)
	// Should contain compound nouns
	compoundFound := false
	for _, k := range keywords {
		if k == "人工知能" || k == "機械学習" || k == "自然言語" || k == "言語処理" {
			compoundFound = true
			break
		}
	}
	if !compoundFound {
		t.Errorf("expected compound nouns, got %v", keywords)
	}
}

func TestExtractJapaneseKeywordsWithFrequency(t *testing.T) {
	text := "テストのテストです。Go言語とGo言語のテスト。"
	freq := ExtractJapaneseKeywordsWithFrequency(text)
	if len(freq) == 0 {
		t.Error("expected some keywords, got none")
	}
}

func TestIsSymbolOrPunctuation(t *testing.T) {
	if !isSymbolOrPunctuation("！＠＃") {
		t.Error("expected true for symbols")
	}
	if isSymbolOrPunctuation("テスト") {
		t.Error("expected false for Japanese text")
	}
	if isSymbolOrPunctuation("abc123") {
		t.Error("expected false for alphanum")
	}
	if isSymbolOrPunctuation("") {
		t.Error("expected false for empty string")
	}
}
