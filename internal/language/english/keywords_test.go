package english

import (
	"testing"
)

func dummyNormalize(word string) string { return word }

func TestNormalizeEnglishKeyword(t *testing.T) {
	pluralMap := map[string]string{"children": "child", "people": "person"}
	invariant := map[string]bool{"news": true, "series": true}

	tests := []struct {
		input    string
		expected string
	}{
		{"cities", "city"},
		{"classes", "class"},      // sses → ss
		{"boxes", "box"},          // xes → x
		{"watches", "watch"},      // ches → ch
		{"cases", "case"},         // ses → se
		{"databases", "database"}, // ses → se
		{"services", "service"},   // ses → se
		{"dogs", "dog"},           // simple s
		{"news", "news"},          // invariant
		{"children", "child"},     // plural map
		{"bus", "bus"},            // ends in "us", no strip
		{"analysis", "analysis"},  // ends in "is", no strip
	}
	for _, tt := range tests {
		got := NormalizeEnglishKeyword(tt.input, pluralMap, invariant)
		if got != tt.expected {
			t.Errorf("NormalizeEnglishKeyword(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestExtractEnglishKeywords(t *testing.T) {
	stopWords := map[string]int{"the": 0, "is": 0}
	text := "The quick brown fox jumps over the lazy dog."
	keywords := ExtractEnglishKeywords(text, stopWords, dummyNormalize)
	found := false
	for _, k := range keywords {
		if k == "quick" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'quick' in keywords, got %v", keywords)
	}
	for _, k := range keywords {
		if k == "the" || k == "is" {
			t.Errorf("stopword '%s' should not be included", k)
		}
	}
}
