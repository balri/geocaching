.PHONY: setup test lint vuln

setup:
	git config core.hooksPath .githooks

test:
	go test ./...

lint:
	golangci-lint run ./...

vuln:
	govulncheck ./...
