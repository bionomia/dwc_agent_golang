# dwc_agent_golang
Ruby-based human name parser, https://github.com/bionomia/dwc_agent ported to Go using Claude

[![Build Status](https://github.com/bionomia/dwc_agent_golang/actions/workflows/go.yml/badge.svg)](https://github.com/bionomia/dwc_agent_golang/actions)

## Run tests
```go
go test ./...
```

## Build the tools
```go
go build -o dwcagent ./cmd/dwcagent
go build -o dwcagent-server ./cmd/dwcagent-server
```

## Run the tools
```go
./dwcagent "13267 (male) W.J. Cody; 13268 (female) W.E. Kemp"
```

## Run the server
```go
./dwcagent-server -port 7654 &
```
See the main.go file in cmd/dwcagent-server for example usage.
