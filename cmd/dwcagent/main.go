// dwcagent parses and cleans a Darwin Core recordedBy / identifiedBy string
// and prints each name as a JSON array. Mirrors the behaviour of the Ruby
// DwcAgent gem: DwcAgent.parse(input).map { |n| DwcAgent.clean(n) }
//
// Names that reduce to DwcAgent.default (all-nil) after cleaning are omitted.
//
// Usage: dwcagent "13267 (male) W.J. Cody; 13268 (female) W.E. Kemp"
package main

import (
	"encoding/json"
	"fmt"
	"os"

	dwcagent "github.com/bionomia/dwc_agent_golang"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: dwcagent <name string>")
		os.Exit(1)
	}
	input := os.Args[1]

	parsed := dwcagent.Parse(input)
	cleaned := make([]dwcagent.Name, 0, len(parsed))
	for _, n := range parsed {
		c := dwcagent.Clean(n)
		if !c.IsDefault() {
			cleaned = append(cleaned, c)
		}
	}

	out, err := json.Marshal(cleaned)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}