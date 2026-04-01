package parser

import (
	"strings"
	"testing"
)

func TestParseHTMLDocument_Success(t *testing.T) {
	html := `<html><head><title>Test</title></head><body><h1>Hello</h1></body></html>`
	doc, err := ParseHTMLDocument(html)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc == nil || doc.Doc == nil {
		t.Fatal("doc or doc.Doc is nil")
	}

	tags := doc.FetchTags("h1")
	if len(tags) != 1 || tags[0] != "Hello" {
		t.Errorf("expected [Hello], got %v", tags)
	}
}

func TestParseHTMLDocument_InvalidHTML(t *testing.T) {
	_, err := ParseHTMLDocument("")
	if err != nil {
		t.Error("goquery should not error on empty string")
	}
}

func TestExtractStructuredText(t *testing.T) {
	html := `<html><head>
	<title>Test Page Title</title>
	<meta name="description" content="A test description">
	<meta name="keywords" content="go, testing, keyword">
	</head><body>
	<script>var x = 1;</script>
	<style>.foo { color: red; }</style>
	<nav>Navigation links</nav>
	<h1>Main Heading</h1>
	<h2>Sub Heading</h2>
	<p>This is the <strong>important</strong> paragraph content.</p>
	<p>Another paragraph with <em>emphasis</em>.</p>
	<a href="/link" title="Link Title">Click here</a>
	<img alt="Test Image" src="test.png">
	<footer>Footer content</footer>
	</body></html>`
	doc, err := ParseHTMLDocument(html)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	segments := doc.ExtractStructuredText()
	if len(segments) == 0 {
		t.Fatal("expected segments, got none")
	}

	// Check title extracted
	foundTitle := false
	for _, s := range segments {
		if s.Source == "title" && s.Weight == 10 && s.Text == "Test Page Title" {
			foundTitle = true
		}
	}
	if !foundTitle {
		t.Error("expected title segment")
	}

	// Check script/style content NOT in segments
	for _, s := range segments {
		if strings.Contains(s.Text, "var x = 1") {
			t.Error("script content should be removed")
		}
		if strings.Contains(s.Text, "color: red") {
			t.Error("style content should be removed")
		}
	}

	// Check h1 extracted
	foundH1 := false
	for _, s := range segments {
		if s.Source == "h1" && s.Text == "Main Heading" {
			foundH1 = true
		}
	}
	if !foundH1 {
		t.Error("expected h1 segment")
	}

	// Check emphasis extracted
	foundEmphasis := false
	for _, s := range segments {
		if s.Source == "emphasis" && (s.Text == "important" || s.Text == "emphasis") {
			foundEmphasis = true
		}
	}
	if !foundEmphasis {
		t.Error("expected emphasis segment")
	}

	// Check meta keywords extracted
	foundMetaKw := false
	for _, s := range segments {
		if s.Source == "meta_keywords" {
			foundMetaKw = true
		}
	}
	if !foundMetaKw {
		t.Error("expected meta_keywords segment")
	}
}
