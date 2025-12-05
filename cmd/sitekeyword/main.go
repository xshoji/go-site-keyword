package main

import (
	"encoding/json"
	"flag"
	"fmt"
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
	commandDescription = "A tool for extracting and analyzing keywords from web pages. Fetches titles, meta tags, and identifies top keywords with their relevance scores."
	// Command options ( the -h, --help option is defined by default in the flag package )
	optionUrl    = defineFlagValue("u", "url" /*    */, Req+"URL" /*   */, "", flag.String, flag.StringVar)
	optionPretty = defineFlagValue("p", "pretty" /* */, "Format JSON output with indentation", false, flag.Bool, flag.BoolVar)
	optionDetail = defineFlagValue("d", "detail" /* */, "Output all details including title and meta tags", false, flag.Bool, flag.BoolVar)
)

func init() {
	// Customize the usage message
	flag.Usage = customUsage(commandDescription)
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
	f := flagFunc(long, defaultValue, flagUsage)
	flagVarFunc(f, short, defaultValue, UsageDummy)
	return f
}

// Custom usage message
func customUsage(description string) func() {
	return func() {
		optionsUsage, requiredOptionExample := getOptionsUsage(false)
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s %s[OPTIONS]\n\n", func() string { e, _ := os.Executable(); return filepath.Base(e) }(), requiredOptionExample)
		fmt.Fprintf(flag.CommandLine.Output(), "Description:\n  %s\n\n", description)
		fmt.Fprintf(flag.CommandLine.Output(), "Options:\n%s", optionsUsage)
	}
}

// Get options usage message
func getOptionsUsage(currentValue bool) (string, string) {
	requiredOptionExample := ""
	optionNameWidth := 0
	usages := make([]string, 0)
	getType := func(v string) string {
		return strings.NewReplacer("*flag.boolValue", "", "*flag.", "<", "Value", ">").Replace(v)
		//return strings.NewReplacer("*flag.boolValue", "", "*flag.", "", "Value", "").Replace(v)
	}
	flag.VisitAll(func(f *flag.Flag) {
		optionNameWidth = max(optionNameWidth, len(fmt.Sprintf("%s %s", f.Name, getType(fmt.Sprintf("%T", f.Value))))+4)
	})
	flag.VisitAll(func(f *flag.Flag) {
		if f.Usage == UsageDummy {
			return
		}
		value := getType(fmt.Sprintf("%T", f.Value))
		if currentValue {
			value = f.Value.String()
		}
		short := strings.Split(f.Usage, UsageDummy)[0]
		mainUsage := strings.Split(f.Usage, UsageDummy)[1]
		if strings.Contains(mainUsage, Req) {
			requiredOptionExample += fmt.Sprintf("--%s %s ", f.Name, value)
		}
		usages = append(usages, fmt.Sprintf("  -%-1s, --%-"+strconv.Itoa(optionNameWidth)+"s %s\n", short, f.Name+" "+value, mainUsage))
	})
	sort.SliceStable(usages, func(i, j int) bool {
		return strings.Count(usages[i], Req) > strings.Count(usages[j], Req)
	})
	return strings.Join(usages, ""), requiredOptionExample
}
