// dwcagent-similarity computes the similarity score between two given-name strings.
//
// Usage: dwcagent-similarity "John C." "John"
package main

import (
	"fmt"
	"os"

	dwcagent "github.com/bionomia/dwc_agent"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: dwcagent-similarity <given1> <given2>")
		os.Exit(1)
	}
	score := dwcagent.SimilarityScore(os.Args[1], os.Args[2])
	fmt.Println(score)
}
