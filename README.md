# cacheodon package

Get package:
```
go get github.com/balri/cacheodon@v0.2.6
```

# Check vulnerabilities
```
export PATH="$PATH:$(go env GOPATH)/bin"
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

# Update vulnerabilities
```
go get -u ./...
go mody tidy
```

# Run linting
```
golangci-lint ./...
```
