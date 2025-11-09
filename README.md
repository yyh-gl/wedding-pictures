# Wedding Pictures Slideshow

結婚式の披露宴で参加者が写真を共有できるシステムです。LINE公式アカウントに写真を送信すると、AWS S3に保存され、会場のモニターでスライドショー表示されます。

## 機能

- LINE公式アカウントから画像アップロード
- AWS S3への自動保存
- リアルタイムスライドショー表示
- 新着画像の優先表示
- すべての画像を表示後、自動的に最初から再生

## アーキテクチャ

- **バックエンド**: Go
- **データベース**: Supabase (PostgreSQL)
- **ストレージ**: AWS S3
- **メッセージング**: LINE Messaging API
- **フロントエンド**: HTML/CSS/JavaScript

## セットアップ

### 1. 必要な準備

#### LINE Bot設定
1. [LINE Developers Console](https://developers.line.biz/)でチャンネルを作成
2. Channel SecretとChannel Access Tokenを取得

#### AWS S3設定
1. [AWS Console](https://console.aws.amazon.com/)でS3バケットを作成
2. バケットのパブリックアクセス設定を有効化（画像の公開読み取り用）
3. IAMユーザーを作成し、S3への読み書き権限を付与
4. アクセスキーIDとシークレットアクセスキーを取得
5. 必要に応じてCORSを設定

#### Supabase設定
1. [Supabase](https://supabase.com/)でプロジェクトを作成
2. Project URLとService Role Keyを取得
3. マイグレーションを実行してテーブルを作成

### 2. 環境変数の設定

`.env.example`をコピーして`.env`を作成し、必要な値を設定します。

```bash
cp .env.example .env
```

`.env`ファイルを編集：

```env
LINE_CHANNEL_SECRET=your_channel_secret
LINE_CHANNEL_ACCESS_TOKEN=your_channel_access_token

SUPABASE_URL=https://your-project.supabase.co
SUPABASE_SERVICE_ROLE_KEY=your_service_role_key

AWS_REGION=ap-northeast-1
AWS_S3_BUCKET=your-bucket-name
AWS_ACCESS_KEY_ID=your_access_key_id
AWS_SECRET_ACCESS_KEY=your_secret_access_key
```

### 3. アプリケーションの起動

Goアプリケーションを起動：

```bash
go run main.go
```

### 4. LINE Webhookの設定

1. [LINE Developers Console](https://developers.line.biz/)でチャンネル設定を開く
2. Webhook URLに `https://your-domain/callback` を設定
3. Webhookを有効化

## 使い方

### 画像のアップロード

LINE公式アカウントに画像を送信すると、自動的にAWS S3に保存されます。

### スライドショーの表示

ブラウザで以下のURLにアクセス：

```
http://localhost:8080/slideshow.html
```

または本番環境のURL：

```
https://your-domain/slideshow.html
```

### スライドショーの仕様

- 各画像は5秒間表示されます
- 新しい画像が追加されると、優先的に表示されます
- 新しい画像には「NEW」バッジが表示されます
- すべての画像を表示し終わったら、最初から再度表示されます
- 10秒ごとに新しい画像をチェックします

## API エンドポイント

### `GET /api/images`

表示する画像のリストを取得します。

**レスポンス例:**
```json
[
  {
    "id": 1,
    "file_url": "https://your-bucket.s3.ap-northeast-1.amazonaws.com/images/...",
    "is_new": true,
    "created_at": "2025-01-01 12:00:00"
  }
]
```

### `POST /api/images/displayed`

画像を表示済みとしてマークします。

**リクエスト例:**
```json
{
  "id": 1
}
```

### `POST /callback`

LINE Webhookのエンドポイント（LINE Platformから呼び出されます）

### `GET /health`

ヘルスチェックエンドポイント

## 開発

### ローカル開発環境

```bash
# 依存関係のインストール
go mod download

# 開発サーバーの起動（ホットリロード付き）
docker-compose up
```

Airによるホットリロードが有効になっており、コードの変更が自動的に反映されます。

### ディレクトリ構造

```
.
├── internal/
│   ├── db/           # データベースモデルと接続
│   ├── storage/      # AWS S3統合
│   └── handler/      # HTTPハンドラー
├── static/
│   └── slideshow.html  # スライドショーUI
├── supabase/
│   └── migrations/   # データベースマイグレーション
├── main.go           # エントリーポイント
└── .env
```

## トラブルシューティング

### 画像が表示されない

1. AWS S3のバケット設定が正しいか確認
2. IAMユーザーの権限が適切か確認
3. S3バケットのパブリックアクセス設定を確認
4. Supabaseのデータベースに画像が保存されているか確認

### LINEからメッセージが届かない

1. Webhook URLが正しく設定されているか確認
2. SSL証明書が有効か確認（LINEはHTTPSのみサポート）
3. サーバーのログを確認して、エラーがないかチェック

## ライセンス

MIT
