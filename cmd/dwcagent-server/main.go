// dwcagent-server exposes the dwc_agent parser over a local HTTP API.
// Intended to run as a sidecar alongside a Spark executor.
//
// POST /parse
//   Body:  {"input": "W.J. Cody; R.D.M. Page"}
//   Reply: [{"family":"Cody","given":"W.J.",...}, ...]
//
// POST /parse_batch
//   Body:  {"inputs": ["W.J. Cody", "R.D.M. Page", ...]}
//   Reply: [{"input":"W.J. Cody","parsed":[...]}, ...]
//
// GET /health  →  200 OK
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	dwcagent "github.com/bionomia/dwc_agent_golang"
)

var flagPort = flag.String("port", "7654", "Port to listen on")

func main() {
	flag.Parse()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/parse", handleParse)
	http.HandleFunc("/parse_batch", handleParseBatch)

	addr := "127.0.0.1:" + *flagPort
	log.Printf("dwcagent-server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// ── /parse ────────────────────────────────────────────────────────────────────

type parseRequest struct {
	Input string `json:"input"`
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req parseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parseAndClean(req.Input))
}

// ── /parse_batch ──────────────────────────────────────────────────────────────

type batchRequest struct {
	Inputs []string `json:"inputs"`
}

type batchItem struct {
	Input  string         `json:"input"`
	Parsed []dwcagent.Name `json:"parsed"`
}

func handleParseBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req batchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	results := make([]batchItem, len(req.Inputs))
	for i, input := range req.Inputs {
		results[i] = batchItem{Input: input, Parsed: parseAndClean(input)}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// ── shared ────────────────────────────────────────────────────────────────────

func parseAndClean(input string) []dwcagent.Name {
	names := dwcagent.Parse(input)
	cleaned := make([]dwcagent.Name, 0, len(names))
	for _, n := range names {
		c := dwcagent.Clean(n)
		if !c.IsDefault() {
			cleaned = append(cleaned, c)
		}
	}
	return cleaned
}
