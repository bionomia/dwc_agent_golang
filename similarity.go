package dwcagent

import (
	"regexp"
	"strings"
)

var dotSpaceRe = regexp.MustCompile(`\.\s+`)

// SimilarityScore computes a similarity score between two given-name strings.
// Logic inspired by R.D.M. Page (orcid.org/0000-0002-7101-9767).
//
// Examples:
//
//	SimilarityScore("John C.", "John Charles") → 2.0
//	SimilarityScore("John C.", "John")         → 1.1
//	SimilarityScore("John C.", "Joshua")       → 0.0
//	SimilarityScore("John C.", "John R.")      → 0.0
func SimilarityScore(given1, given2 string) float64 {
	given1 = dotSpaceRe.ReplaceAllString(strings.TrimSpace(given1), ".")
	given2 = dotSpaceRe.ReplaceAllString(strings.TrimSpace(given2), ".")

	g1arr := splitGiven(given1)
	g2arr := splitGiven(given2)

	largest, smallest := g1arr, g2arr
	if len(g2arr) > len(g1arr) {
		largest, smallest = g2arr, g1arr
	}

	score := 0.0
	for i, val := range largest {
		if i < len(smallest) {
			smallVal := smallest[i]
			if len(val) == 0 || len(smallVal) == 0 {
				return 0
			}
			if val[0] != smallVal[0] {
				return 0
			}
			if len(val) > 1 && len(smallVal) > 1 && !strings.Contains(val, smallVal) {
				return 0
			}
			score += 1.0
		} else {
			score += 0.1
		}
	}
	return score
}

var splitGivenRe = regexp.MustCompile(`[.\s]`)

func splitGiven(s string) []string {
	var tokens []string
	for _, part := range splitGivenRe.Split(s, -1) {
		if part != "" {
			tokens = append(tokens, part)
		}
	}
	return tokens
}
