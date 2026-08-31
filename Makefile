VERSION ?= 2.0.3
DIST := dist

.PHONY: test test-race vet build clean

test:
	go test ./... -count=1

test-race:
	go test -race ./... -count=1

vet:
	go vet ./...

build: test vet
	rm -rf $(DIST)
	mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o $(DIST)/commit-ai-linux-amd64 ./cmd/commit-ai
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o $(DIST)/commit-ai-linux-arm64 ./cmd/commit-ai
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o $(DIST)/commit-ai-darwin-amd64 ./cmd/commit-ai
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o $(DIST)/commit-ai-darwin-arm64 ./cmd/commit-ai
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o $(DIST)/commit-ai-windows-amd64.exe ./cmd/commit-ai
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o $(DIST)/commit-ai-windows-arm64.exe ./cmd/commit-ai

clean:
	rm -rf $(DIST)
