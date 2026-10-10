# Stockbox
後で読むURLのストックと共有をするシステムです。

## Getting Started
sqlite3　インストール
sqlic インストール


```bash
# Gitからクローン
git clone https://github.com/tatsuun08/Stockbox.git
cd Stockbox

# sqlc/sqliteのインストールが必要
sqlc generate

# 必要なパッケージをインストール
go mod tidy
go run main.go

# 別ターミナルで
curl localhost:1323/test # {"message":"THIS IS TEST!!"}
```
## TODO
1. 登録・ログイン
		普通のメールアドレス・パスワード
2. 自分のボックス画面
		上にURLを貼る入力欄と登録ボタン
		貼ってボタンを押すと取得中（非同期処理）
		数秒後にタイトルと要約文が勝手に表示される
3. チームのボックス画面
		自分が公開にしたブックマークだけが、他人のログインしている人にも見える
		他人の公開したものも見れる 
