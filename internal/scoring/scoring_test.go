package scoring

import "testing"

func TestRankKeywordsByScore(t *testing.T) {
	scoreMap := map[string]int{
		"go":    10,
		"python": 5,
		"java":  7,
	}
	originalMap := map[string]string{
		"go":    "Go",
		"python": "Python",
		"java":  "Java",
	}

	result := RankKeywordsByScore(scoreMap, originalMap, 2)
	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}
	if result[0].Keyword != "Go" || result[0].Score != 10 {
		t.Errorf("expected first to be Go(10), got %v", result[0])
	}
	if result[1].Keyword != "Java" || result[1].Score != 7 {
		t.Errorf("expected second to be Java(7), got %v", result[1])
	}
}

func TestRankKeywordsByScore_LimitZero(t *testing.T) {
	scoreMap := map[string]int{"a": 1, "b": 2}
	originalMap := map[string]string{"a": "A", "b": "B"}
	result := RankKeywordsByScore(scoreMap, originalMap, 0)
	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}
}

func TestCalculatePositionBoost(t *testing.T) {
	if CalculatePositionBoost(0.1) != 1.5 {
		t.Error("expected 1.5 for position 0.1")
	}
	if CalculatePositionBoost(0.5) != 1.2 {
		t.Error("expected 1.2 for position 0.5")
	}
	if CalculatePositionBoost(0.8) != 1.0 {
		t.Error("expected 1.0 for position 0.8")
	}
}

func TestCalculateLengthPenalty(t *testing.T) {
	if CalculateLengthPenalty("a") != 0.3 {
		t.Error("expected 0.3 for 1-char")
	}
	if CalculateLengthPenalty("ab") != 0.7 {
		t.Error("expected 0.7 for 2-char")
	}
	if CalculateLengthPenalty("abc") != 1.0 {
		t.Error("expected 1.0 for 3-char")
	}
	// Japanese 2-char
	if CalculateLengthPenalty("日本") != 0.7 {
		t.Error("expected 0.7 for 2-rune Japanese")
	}
}

func TestRankKeywordsAdvanced(t *testing.T) {
	scoreMap := map[string]float64{
		"golang": 20.0,
		"go":     15.0,
		"test":   10.0,
		"x":      5.0,
	}
	originalMap := map[string]string{
		"golang": "Golang",
	}
	firstPosMap := map[string]float64{
		"golang": 0.1,
		"go":     0.5,
		"test":   0.8,
		"x":      0.1,
	}
	result := RankKeywordsAdvanced(scoreMap, originalMap, firstPosMap, 3, 2.0)
	if len(result) > 3 {
		t.Errorf("expected at most 3 results, got %d", len(result))
	}
	if len(result) == 0 {
		t.Fatal("expected some results")
	}
	if result[0].Keyword != "Golang" {
		t.Errorf("expected first keyword to be 'Golang', got '%s'", result[0].Keyword)
	}
}
