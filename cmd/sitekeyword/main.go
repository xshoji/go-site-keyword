package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xshoji/go-site-keyword/pkg/analyzer"
	"github.com/xshoji/go-site-keyword/pkg/config"
)

const (
	UsageRequiredPrefix = "\u001B[33m" + "(REQ)" + "\u001B[0m "
	UsageDummy          = "########"
	TimeFormat          = "2006-01-02 15:04:05.0000 [MST]"
)

var (
	// Command options ( the -h, --help option is defined by default in the flag package )
	commandDescription     = "A tool for extracting and analyzing keywords from web pages. \n  Fetches titles, meta tags, and identifies top keywords with their relevance scores."
	commandOptionMaxLength = 0
	optionUrl              = defineFlagValue("u", "url" /*    */, UsageRequiredPrefix+"URL" /*   */, "").(*string)
	optionPretty           = defineFlagValue("p", "pretty" /* */, "Format JSON output with indentation", false).(*bool)
	optionDetail           = defineFlagValue("d", "detail" /* */, "Output all details including title and meta tags", false).(*bool)
)

func init() {
	formatUsage(commandDescription, &commandOptionMaxLength, new(bytes.Buffer))
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

// convertKeywords 関数は不要になったため削除

// Helper function for flag
func defineFlagValue(short, long, description string, defaultValue any) (f any) {
	flagUsage := short + UsageDummy + description
	switch v := defaultValue.(type) {
	case string:
		f = flag.String(short, "", UsageDummy)
		flag.StringVar(f.(*string), long, v, flagUsage)
	case int:
		f = flag.Int(short, 0, UsageDummy)
		flag.IntVar(f.(*int), long, v, flagUsage)
	case bool:
		f = flag.Bool(short, false, UsageDummy)
		flag.BoolVar(f.(*bool), long, v, flagUsage)
	case float64:
		f = flag.Float64(short, 0.0, UsageDummy)
		flag.Float64Var(f.(*float64), long, v, flagUsage)
	default:
		panic("unsupported flag type")
	}
	return
}

func formatUsage(description string, maxLength *int, buffer *bytes.Buffer) {
	func() { flag.CommandLine.SetOutput(buffer); flag.Usage(); flag.CommandLine.SetOutput(os.Stderr) }()
	usageOption := regexp.MustCompile("(-\\S+)( *\\S*)+\n*\\s+"+UsageDummy+"\n\\s*").ReplaceAllString(buffer.String(), "")
	re := regexp.MustCompile("\\s(-\\S+)( *\\S*)( *\\S*)+\n\\s+(.+)")
	usageFirst := strings.Replace(strings.Replace(strings.Split(usageOption, "\n")[0], ":", " [OPTIONS] [-h, --help]", -1), " of ", ": ", -1) + "\n\nDescription:\n  " + description + "\n\nOptions:\n"
	usageOptions := re.FindAllString(usageOption, -1)
	for _, v := range usageOptions {
		*maxLength = max(*maxLength, len(re.ReplaceAllString(v, " -$1")+re.ReplaceAllString(v, "$2"))+2)
	}
	usageOptionsRep := make([]string, 0)
	for _, v := range usageOptions {
		usageOptionsRep = append(usageOptionsRep, fmt.Sprintf("  -%-1s,%-"+strconv.Itoa(*maxLength)+"s%s", strings.Split(re.ReplaceAllString(v, "$4"), UsageDummy)[0], re.ReplaceAllString(v, " -$1")+re.ReplaceAllString(v, "$2"), strings.Split(re.ReplaceAllString(v, "$4"), UsageDummy)[1]+"\n"))
	}
	sort.SliceStable(usageOptionsRep, func(i, j int) bool {
		return strings.Count(usageOptionsRep[i], UsageRequiredPrefix) > strings.Count(usageOptionsRep[j], UsageRequiredPrefix)
	})
	flag.Usage = func() { _, _ = fmt.Fprint(flag.CommandLine.Output(), usageFirst+strings.Join(usageOptionsRep, "")) }
}
