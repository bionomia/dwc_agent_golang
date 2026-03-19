# dwc_agent_golang
DwC Agent parser ported to Go using Claude

## Run tests
go test ./...

## Build the tools
go build -o dwcagent ./cmd/dwcagent
go build -o dwcagent-server ./cmd/dwcagent-server

## Run the tools
./dwcagent "13267 (male) W.J. Cody; 13268 (female) W.E. Kemp"

## Run the server
./dwcagent-server -port 7654 &