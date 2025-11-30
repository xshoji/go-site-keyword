package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xshoji/go-site-keyword/pkg/analyzer"
	"github.com/xshoji/go-site-keyword/pkg/config"
)

const (
	Req        = "\x1b[33m" + "(required)" + "\x1b[0m "
	UsageDummy = "########"
	TimeFormat = "2006-01-02 15:04:05.0000 [MST]"
)

var (
	commandDescription           = "A tool for extracting and analyzing keywords from web pages. Fetches titles, meta tags, and identifies top keywords with their relevance scores."
	commandOptionMaxLength       = 0
	commandRequiredOptionExample = "" // Auto-adjusted in defineFlagValue
	// Command options ( the -h, --help option is defined by default in the flag package )
	optionUrl    = defineFlagValue("u", "url" /*    */, Req+"URL" /*   */, "", flag.String, flag.StringVar)
	optionPretty = defineFlagValue("p", "pretty" /* */, "Format JSON output with indentation", false, flag.Bool, flag.BoolVar)
	optionDetail = defineFlagValue("d", "detail" /* */, "Output all details including title and meta tags", false, flag.Bool, flag.BoolVar)
)

func init() {
	// Customize the usage message
	flag.Usage = customUsage(os.Stdout, commandDescription, strconv.Itoa(commandOptionMaxLength), commandRequiredOptionExample)
}

// Build:
// $ GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -trimpath ./cmd/sitekeyword
func main() {
	flag.Parse()
	if *optionUrl == "" {
		flag.Usage()
		os.Exit(0)
	}

	cfg := config.DefaultConfig()
	anlz, err := analyzer.NewAnalyzer(*optionUrl, cfg)
	if err != nil {
		handleError(err, "NewAnalyzer")
		os.Exit(1)
	}
	// 解析結果を取得
	result, err := anlz.GetAnalysisResult(20)
	if err != nil {
		handleError(err, "GetAnalysisResult")
		os.Exit(1)
	}

	// JSON形式で出力
	var jsonData []byte
	var outputObj interface{}

	// 詳細表示フラグに応じて出力形式を切り替え
	if *optionDetail {
		// 詳細表示：全てのフィールドを出力
		outputObj = result
	} else {
		// デフォルト：keywordsのみを出力（匿名構造体を使用）
		outputObj = struct {
			Keywords interface{} `json:"keywords"`
		}{
			Keywords: result.Keywords,
		}
	}

	if *optionPretty {
		// インデント付きのJSON出力（pretty形式）
		jsonData, err = json.MarshalIndent(outputObj, "", "  ")
	} else {
		// 1行のJSON出力
		jsonData, err = json.Marshal(outputObj)
	}

	if err != nil {
		handleError(err, "JSON Marshal")
		os.Exit(1)
	}
	fmt.Println(string(jsonData))
}

// =======================================
// Common Utils
// =======================================

func handleError(err error, prefixErrMessage string) {
	if err != nil {
		fmt.Printf("%s [ERROR %s]: %v\n", time.Now().Format(TimeFormat), prefixErrMessage, err)
	}
}

// =======================================
// flag Utils
// =======================================

// Helper function for flag
func defineFlagValue[T comparable](short, long, description string, defaultValue T, flagFunc func(name string, value T, usage string) *T, flagVarFunc func(p *T, name string, value T, usage string)) *T {
	flagUsage := short + UsageDummy + description
	var zero T
	if defaultValue != zero {
		flagUsage = flagUsage + fmt.Sprintf(" (default %v)", defaultValue)
	}
	if strings.Contains(description, Req) {
		commandRequiredOptionExample = commandRequiredOptionExample + fmt.Sprintf("--%s %T ", long, defaultValue)
	}
	commandOptionMaxLength = max(commandOptionMaxLength, len(long)+12)
	f := flagFunc(long, defaultValue, flagUsage)
	flagVarFunc(f, short, defaultValue, UsageDummy)
	return f
}

// Custom usage message
func customUsage(output io.Writer, description, fieldWidth, requiredOptionExample string) func() {
	return func() {
		fmt.Fprintf(output, "Usage: %s %s[OPTIONS]\n\n", func() string { e, _ := os.Executable(); return filepath.Base(e) }(), requiredOptionExample)
		fmt.Fprintf(output, "Description:\n  %s\n\n", description)
		fmt.Fprintf(output, "Options:\n%s", getOptionsUsage(fieldWidth, false))
	}
}

// Get options usage message
func getOptionsUsage(fieldWidth string, currentValue bool) string {
	optionUsages := make([]string, 0)
	flag.VisitAll(func(f *flag.Flag) {
		if f.Usage == UsageDummy {
			return
		}
		value := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(fmt.Sprintf("%T", f.Value), "*flag.", ""), "Value", ""), "bool", "")
		if currentValue {
			value = f.Value.String()
		}
		format := "  -%-1s, --%-" + fieldWidth + "s %s\n"
		short := strings.Split(f.Usage, UsageDummy)[0]
		mainUsage := strings.Split(f.Usage, UsageDummy)[1]
		optionUsages = append(optionUsages, fmt.Sprintf(format, short, f.Name+" "+value, mainUsage))
	})
	sort.SliceStable(optionUsages, func(i, j int) bool {
		return strings.Count(optionUsages[i], Req) > strings.Count(optionUsages[j], Req)
	})
	return strings.Join(optionUsages, "")
}
