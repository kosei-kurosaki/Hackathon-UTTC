# Dockerfile

# ベースイメージとしてGolangイメージを使用
FROM golang:1.18 as build

# 作業ディレクトリを設定
WORKDIR /app

# ソースコードをコピー
COPY . .

# ビルド
RUN go mod download

Run go build -v main.go

# コンテナ起動時に実行するコマンドを指定
CMD["app/main.go"]
