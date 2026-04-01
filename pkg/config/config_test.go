package config

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Timeout != 10*time.Second {
		t.Errorf("expected Timeout 10s, got %v", cfg.Timeout)
	}
	if cfg.UserAgent == "" {
		t.Error("UserAgent should not be empty")
	}
	if cfg.ScoreWeights.Title != 10 {
		t.Errorf("expected Title weight 10, got %d", cfg.ScoreWeights.Title)
	}
	if cfg.ScoreWeights.MetaKeyword != 8 {
		t.Errorf("expected MetaKeyword weight 8, got %d", cfg.ScoreWeights.MetaKeyword)
	}
	if cfg.ScoreWeights.Body != 1 {
		t.Errorf("expected Body weight 1, got %d", cfg.ScoreWeights.Body)
	}
	if cfg.MaxKeywords != 20 {
		t.Errorf("expected MaxKeywords 20, got %d", cfg.MaxKeywords)
	}
	if cfg.MinScore != 2.0 {
		t.Errorf("expected MinScore 2.0, got %f", cfg.MinScore)
	}
}
