.PHONY: all

all: ssl

ssl:
	go mod tidy
	go vet ./cmd/ssl
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X main.buildTime=`date +%Y%m%d.%H:%M:%S` -X main.buildCommit=`git rev-parse --short=12 HEAD` -X main.buildBranch=`git branch --show-current`" -o ./releases/tencent_ssl_darwin_amd64 ./cmd/ssl
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X main.buildTime=`date +%Y%m%d.%H:%M:%S` -X main.buildCommit=`git rev-parse --short=12 HEAD` -X main.buildBranch=`git branch --show-current`" -o ./releases/tencent_ssl_darwin_arm64 ./cmd/ssl
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.buildTime=`date +%Y%m%d.%H:%M:%S` -X main.buildCommit=`git rev-parse --short=12 HEAD` -X main.buildBranch=`git branch --show-current`" -o ./releases/tencent_ssl_linux_amd64 ./cmd/ssl
	GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -X main.buildTime=`date +%Y%m%d.%H:%M:%S` -X main.buildCommit=`git rev-parse --short=12 HEAD` -X main.buildBranch=`git branch --show-current`" -o ./releases/tencent_ssl_linux_arm64 ./cmd/ssl
