<div align="center">
  <h1>📝 LLMO Analysis</h1>
  <!-- 必要であればヘッダー画像を追加 -->
  <!-- <img src="path/to/header.png" alt="Header Image"> -->
</div>

<div align="center">
  <!-- 技術スタックのバッジを追加 -->
  <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <!-- 他の技術バッジも追加 (例: Gemini API) -->
</div>

## 📖 概要

このプロジェクトは、指定されたクエリリストを Gemini API に送信し、その応答内容を分析するツールです。
応答に特定の会社名、製品名、ドメイン名が含まれているかをチェックし、結果をタイムスタンプと共に CSV ファイルに出力します。

## 📚 参考記事

このツールの背景や詳細については、以下の Zenn 記事もご参照ください。

- [Web 開発者のための LLMO:用語解説から効果測定まで](https://zenn.dev/tkwbr999/articles/65c0a1a0ba8f1d)

## 🛠️ セットアップ

### 前提条件

- Go 1.x 以降がインストールされていること
- Gemini API キーを取得済みであること

### インストール

1.  リポジトリをクローンします:
    ```bash
    git clone https://github.com/tKwbr999/llmo-analysis.git
    cd llmo-analysis
    ```
2.  依存関係をインストールします:
    ```bash
    go mod download
    ```

### 環境設定

1.  `.env.example` ファイルをコピーして `.env` ファイルを作成します:
    ```bash
    cp .env.example .env
    ```
2.  `.env` ファイルを開き、以下の環境変数を設定します:
    - `GEMINI_API_KEY`: あなたの Gemini API キー
    - `GEMINI_API_MODEL`: 使用する Gemini モデル名 (例: "gemini-1.5-flash")
    - `DOMAIN_NAME`: 応答内でチェックしたいドメイン名 (例: "example.com")
    - `COMPANY_NAME`: 応答内でチェックしたい会社名 (例: "Example Inc.")
    - `PRODUCT_NAMES`: チェックしたい製品名のリスト (カンマ区切り、例: "ProductA,ProductB,ServiceC")
    - `TARGET_QUERIES`: Gemini API に送信するクエリのリスト (パイプ `|` 区切り、例: "製品 A について教えて|サービス C の料金は？")

## 🚀 使い方

以下のコマンドを実行します:

```bash
go run cmd/main.go
```

実行が完了すると、プロジェクトルートに `llmo_monitoring_YYYYMMDD.csv` という名前の CSV ファイルが生成されます。
また、コンソールには各指標の言及率が表示されます。

## 📄 出力 CSV フォーマット

CSV ファイルには以下の列が含まれます:

- `query`: 送信したクエリ
- `timestamp`: 処理実行時のタイムスタンプ (RFC3339 形式)
- `company_mentioned`: 会社名が言及されたか (true/false)
- `products_mentioned`: 言及された製品名のリスト (カンマ区切り、なければ "なし")
- `url_mentioned`: ドメイン名が言及されたか (true/false)
- `full_response`: Gemini API からの完全な応答テキスト

## 📁 プロジェクト構造

```
.
├── .env.example      # 環境変数設定例
├── .gitignore        # Git 無視ファイル
├── README.md         # このファイル
├── cmd/              # コマンドラインエントリポイント
│   └── main.go       # メインプログラム
├── go.mod            # Go モジュール定義
└── go.sum            # Go モジュールチェックサム
```

(必要に応じて他のセクションを追加)
