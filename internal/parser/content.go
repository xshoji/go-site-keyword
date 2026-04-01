package parser

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// HTMLDocument は goquery.Document のラッパー
// 今後の拡張やテスト容易性のための型
// （必要に応じて拡張可）
type HTMLDocument struct {
	Doc *goquery.Document
}

// ParseHTMLDocument はHTML文字列から goquery.Document を生成します
func ParseHTMLDocument(html string) (*HTMLDocument, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("Failed to parse HTML document: %w", err)
	}
	return &HTMLDocument{Doc: doc}, nil
}

// FetchTags 指定タグの内容をすべて抜き出します
func (h *HTMLDocument) FetchTags(tag string) []string {
	var result []string
	h.Doc.Find(tag).Each(func(i int, s *goquery.Selection) {
		result = append(result, s.Text())
	})
	return result
}

// TextSegment represents extracted text with its source type and weight
type TextSegment struct {
	Text   string
	Source string // "title", "meta_keywords", "meta_description", "h1", "h2", "h3", "emphasis", "anchor", "alt", "body"
	Weight int
}

// ExtractStructuredText extracts text from multiple HTML sources with noise removal.
// It removes script/style/noscript/svg/iframe tags, then extracts text from
// various elements with appropriate weights.
func (h *HTMLDocument) ExtractStructuredText() []TextSegment {
	var segments []TextSegment

	// Clone the document to avoid mutating the original
	clone := h.Doc.Clone()

	// Remove noise elements from the clone
	clone.Find("script, style, noscript, svg, iframe, nav, footer, header, aside").Remove()
	// Remove common site-specific noise (sidebars, navigation boxes, info panels)
	clone.Find("[role='navigation'], [role='complementary'], [role='banner']").Remove()
	clone.Find(".sidebar, .navbox, .infobox, .mw-panel, .interlanguage-link, .mw-editsection").Remove()
	clone.Find("[class*='sidebar'], [class*='navbox'], [class*='infobox']").Remove()

	// title
	if text := strings.TrimSpace(clone.Find("title").Text()); text != "" {
		segments = append(segments, TextSegment{Text: text, Source: "title", Weight: 10})
	}

	// meta[name=keywords]
	if content, exists := clone.Find(`meta[name=keywords]`).Attr("content"); exists {
		for _, kw := range strings.Split(content, ",") {
			if trimmed := strings.TrimSpace(kw); trimmed != "" {
				segments = append(segments, TextSegment{Text: trimmed, Source: "meta_keywords", Weight: 8})
			}
		}
	}

	// meta[name=description] or fallback to og:description
	if content, exists := clone.Find(`meta[name=description]`).Attr("content"); exists {
		if trimmed := strings.TrimSpace(content); trimmed != "" {
			segments = append(segments, TextSegment{Text: trimmed, Source: "meta_description", Weight: 6})
		}
	} else if content, exists := clone.Find(`meta[property="og:description"]`).Attr("content"); exists {
		if trimmed := strings.TrimSpace(content); trimmed != "" {
			segments = append(segments, TextSegment{Text: trimmed, Source: "meta_description", Weight: 6})
		}
	}

	// Headings
	clone.Find("h1").Each(func(i int, s *goquery.Selection) {
		if text := strings.TrimSpace(s.Text()); text != "" {
			segments = append(segments, TextSegment{Text: text, Source: "h1", Weight: 8})
		}
	})
	clone.Find("h2").Each(func(i int, s *goquery.Selection) {
		if text := strings.TrimSpace(s.Text()); text != "" {
			segments = append(segments, TextSegment{Text: text, Source: "h2", Weight: 5})
		}
	})
	clone.Find("h3").Each(func(i int, s *goquery.Selection) {
		if text := strings.TrimSpace(s.Text()); text != "" {
			segments = append(segments, TextSegment{Text: text, Source: "h3", Weight: 5})
		}
	})

	// Emphasis: strong, em, b
	clone.Find("strong, em, b").Each(func(i int, s *goquery.Selection) {
		if text := strings.TrimSpace(s.Text()); text != "" {
			segments = append(segments, TextSegment{Text: text, Source: "emphasis", Weight: 3})
		}
	})

	// Anchor text
	clone.Find("a").Each(func(i int, s *goquery.Selection) {
		if text := strings.TrimSpace(s.Text()); text != "" {
			segments = append(segments, TextSegment{Text: text, Source: "anchor", Weight: 2})
		}
	})

	// img[alt] — skip placeholder/missing alt text
	clone.Find("img[alt]").Each(func(i int, s *goquery.Selection) {
		if alt, exists := s.Attr("alt"); exists {
			trimmed := strings.TrimSpace(alt)
			lower := strings.ToLower(trimmed)
			if trimmed != "" && lower != "image" && lower != "photo" &&
				!strings.Contains(lower, "missing") && !strings.Contains(lower, "placeholder") {
				segments = append(segments, TextSegment{Text: trimmed, Source: "alt", Weight: 2})
			}
		}
	})

	// a[title]
	clone.Find("a[title]").Each(func(i int, s *goquery.Selection) {
		if title, exists := s.Attr("title"); exists {
			if trimmed := strings.TrimSpace(title); trimmed != "" {
				segments = append(segments, TextSegment{Text: trimmed, Source: "alt", Weight: 2})
			}
		}
	})

	// Body text: p, li, td, dd, blockquote, article
	clone.Find("p, li, td, dd, blockquote, article > *:not(h1):not(h2):not(h3):not(strong):not(em):not(b)").Each(func(i int, s *goquery.Selection) {
		if text := strings.TrimSpace(s.Text()); text != "" {
			segments = append(segments, TextSegment{Text: text, Source: "body", Weight: 1})
		}
	})

	return segments
}
