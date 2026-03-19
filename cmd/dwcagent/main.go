// dwcagent parses a Darwin Core recordedBy / identifiedBy string and prints
// each name as a JSON array.
//
// Usage: dwcagent "13267 (male) W.J. Cody; 13268 (female) W.E. Kemp"
package main

import (
	"encoding/json"
	"fmt"
	"os"

	dwcagent "github.com/bionomia/dwc_agent"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: dwcagent <name string>")
		os.Exit(1)
	}
	input := os.Args[1]
	names := dwcagent.Parse(input)
	out, err := json.Marshal(names)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}
