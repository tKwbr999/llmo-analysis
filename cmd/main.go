package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log" // Added for error logging
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai" // Added Gemini SDK
	"github.com/joho/godotenv"                 // Added for .env file handling
	"google.golang.org/api/option"             // Added for API key option
)

// 結果保存用の構造体
type QueryResult struct {
	Query             string
	Timestamp         string
	CompanyMentioned  bool
	ProductsMentioned string
	URLMentioned      bool
	FullResponse      string
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	geminiModelName := os.Getenv("GEMINI_API_MODEL")
	domainName := os.Getenv("DOMAIN_NAME")
	companyName := os.Getenv("COMPANY_NAME")
	productNames := strings.Split(os.Getenv("PRODUCT_NAMES"), ",")
	targetQueries := strings.Split(os.Getenv("TARGET_QUERIES"), "|")

	if geminiAPIKey == "" || geminiModelName == "" || domainName == "" {
		log.Fatal("環境変数 GEMINI_API_KEY, GEMINI_API_MODEL, DOMAIN_NAME のいずれかが設定されていません。")
	}

	// Gemini クライアントの初期化
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(geminiAPIKey))
	if err != nil {
		log.Fatalf("Gemini クライアントの作成に失敗しました: %v", err)
	}
	defer client.Close()

	// 固定の設定値

	if companyName == "" || len(productNames) == 0 || len(targetQueries) == 0 {
		log.Fatal("環境変数 COMPANY_NAME, PRODUCT_NAMES, TARGET_QUERIES のいずれかが設定されていません。")
	}

	// 結果保存用スライス
	var results []QueryResult

	// 各クエリをテスト
	for _, query := range targetQueries {
		// APIリクエスト (Gemini API を使用)
		answerText, err := callGeminiAPI(ctx, client, geminiModelName, query) // 関数呼び出しを修正
		if err != nil {
			log.Printf("クエリ '%s' の処理中にエラーが発生しました: %v\n", query, err) // エラーログに変更
			continue
		}

		// 会社名/製品名の言及をチェック
		companyMentioned := strings.Contains(strings.ToLower(answerText), strings.ToLower(companyName))
		var mentionedProducts []string
		for _, product := range productNames {
			if strings.Contains(strings.ToLower(answerText), strings.ToLower(product)) {
				mentionedProducts = append(mentionedProducts, product)
			}
		}

		// 製品名の結合
		productsMentioned := strings.Join(mentionedProducts, ", ")
		if productsMentioned == "" {
			productsMentioned = "なし"
		}

		// URLの言及をチェック
		urlMentioned := strings.Contains(strings.ToLower(answerText), strings.ToLower(domainName))

		// 結果を記録
		result := QueryResult{
			Query:             query,
			Timestamp:         time.Now().Format(time.RFC3339),
			CompanyMentioned:  companyMentioned,
			ProductsMentioned: productsMentioned,
			URLMentioned:      urlMentioned,
			FullResponse:      answerText,
		}
		results = append(results, result)

		// レート制限対策
		time.Sleep(1 * time.Second)
	}

	// 結果をCSVに保存
	saveResultsToCSV(results)

	// 集計結果を表示
	printSummary(results)
}

// Gemini APIを呼び出す関数
func callGeminiAPI(ctx context.Context, client *genai.Client, modelName, query string) (string, error) {
	// タイムアウト設定 (必要に応じて調整)
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second) // タイムアウトを60秒に延長
	defer cancel()

	model := client.GenerativeModel(modelName)
	resp, err := model.GenerateContent(ctx, genai.Text(query))
	if err != nil {
		return "", fmt.Errorf("gemini API 呼び出しエラー: %w", err)
	}

	// レスポンスからテキストを抽出
	// Candidates が空、または最初の Candidate の Content が nil、または Parts が空の場合のエラーハンドリング
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		// レスポンスの内容をログに出力して詳細を確認できるようにする
		respJSON, _ := json.MarshalIndent(resp, "", "  ")
		log.Printf("Gemini API から予期しないレスポンスを受け取りました:\n%s", string(respJSON))
		return "", fmt.Errorf("gemini API から有効な回答が得られませんでした")
	}

	// Parts の最初の要素が Text であることを期待
	// (より堅牢にするには、Part の型をチェックすることも検討)
	if textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		return string(textPart), nil
	}

	return "", fmt.Errorf("gemini API レスポンスの最初の Part がテキストではありませんでした")
}

// 結果をCSVに保存する関数
func saveResultsToCSV(results []QueryResult) error {
	filename := fmt.Sprintf("llmo_monitoring_%s.csv", time.Now().Format("20060102"))
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// ヘッダー行を書き込み
	header := []string{"query", "timestamp", "company_mentioned", "products_mentioned", "url_mentioned", "full_response"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// データ行を書き込み
	for _, result := range results {
		companyMentioned := "false"
		if result.CompanyMentioned {
			companyMentioned = "true"
		}

		urlMentioned := "false"
		if result.URLMentioned {
			urlMentioned = "true"
		}

		row := []string{
			result.Query,
			result.Timestamp,
			companyMentioned,
			result.ProductsMentioned,
			urlMentioned,
			result.FullResponse,
		}

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	fmt.Printf("Results saved to %s\n", filename)
	return nil
}

// 集計結果を表示する関数
func printSummary(results []QueryResult) {
	if len(results) == 0 {
		fmt.Println("No results to summarize")
		return
	}

	// 各指標のカウント
	companyMentionCount := 0
	productMentionCount := 0
	urlMentionCount := 0

	for _, result := range results {
		if result.CompanyMentioned {
			companyMentionCount++
		}
		if result.ProductsMentioned != "なし" {
			productMentionCount++
		}
		if result.URLMentioned {
			urlMentionCount++
		}
	}

	// 割合の計算
	totalQueries := len(results)
	companyMentionRate := float64(companyMentionCount) / float64(totalQueries) * 100
	productMentionRate := float64(productMentionCount) / float64(totalQueries) * 100
	urlMentionRate := float64(urlMentionCount) / float64(totalQueries) * 100

	// 結果表示
	fmt.Printf("会社名言及率: %.1f%%\n", companyMentionRate)
	fmt.Printf("製品言及率: %.1f%%\n", productMentionRate)
	fmt.Printf("URL言及率: %.1f%%\n", urlMentionRate)
}
