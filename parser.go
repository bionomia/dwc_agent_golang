package dwcagent

import (
	"regexp"
	"strings"
)

var extraSpaceRe = regexp.MustCompile(`\s{2,}`)

// Parse cleanses the input string and returns a slice of parsed Names.
// Mirrors DwcAgent.parse in the Ruby gem.
func Parse(input string) []Name {
	if strings.TrimSpace(input) == "" {
		return []Name{}
	}

	s := input

	// 1. Strip noise tokens.
	s = stripOut(s)

	// 2. Tidy remaining content.
	s = postStripTidy(s)

	// 3. Apply character substitutions (turns many separators into " | ").
	s = applyCharSubs(s)

	// 4. Apply complex separator substitutions per pipe-segment.
	s = processComplexSeps(s)

	// 5. Remove residual trailing commas/semicolons.
	s = residualTerminatorsRe.ReplaceAllString(s, "")

	// 6. Collapse whitespace.
	s = extraSpaceRe.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)

	if s == "" {
		return []Name{}
	}

	return parseNames(s)
}

// processComplexSeps applies complex separator transforms on each pipe-segment.
func processComplexSeps(s string) string {
	segments := splitByPipeRe.Split(s, -1)
	out := make([]string, 0, len(segments))
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		seg = applyComplexSeparators(seg)
		out = append(out, seg)
	}
	return strings.Join(out, " | ")
}
