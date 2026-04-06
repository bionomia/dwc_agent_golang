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

	// 3b. Convert space-dash-space to " | " — mirrors Ruby's SPLIT_BY \s+-\s+.
	//     Done after charSubs so "-jr"/"-Jr" suffix fixes have already fired.
	s = spaceDashSpaceRe.ReplaceAllString(s, " | ")

	// 4. Apply complex separator substitutions per pipe-segment.
	//    The complex patterns handle cases like "J. & K. Smith" (shared family)
	//    by consuming the "&" before step 4b fires.
	s = processComplexSeps(s)

	// 4b. Convert any remaining " & " to " | ".
	s = strings.ReplaceAll(s, " & ", " | ")

	// 4c. Split on conjunction words.
	s = conjunctionSepRe.ReplaceAllString(s, " | ")

	// 4c2. Re-apply complex separators after conjunction split, so patterns like
	//      "N. Navarro, G. Gómez" (created when 'y' split off 'A Ferreira') are
	//      detected as display-order comma lists.
	s = processComplexSeps(s)

	// 4d. Split on role/verb phrases mid-string.
	s = splitByVerbRe.ReplaceAllString(s, " | ")

	// 4e. Split on remaining punctuation separators.
	s = splitByPunctuationRe.ReplaceAllString(s, " | ")

	// 4f. Strip leading verb phrases from each pipe-segment.
	//     "via Serena Lowartz" → "Serena Lowartz"
	//     "by P. Zika" → "P. Zika"
	//     "prep. C.J. Guiguet" → "C.J. Guiguet"
	//     These appear at the start of a segment after splitting (or as the
	//     entire input), and should leave a single name rather than two.
	s = stripLeadingVerbsPerSegment(s)

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

// stripLeadingVerbsPerSegment strips leading and trailing verb/role phrases from
// each pipe-delimited segment.
// Leading: "via", "by", "prep.", "annotated" → leaves the name that follows
// Trailing: "checked", "annotated", "verified" → left after colon-split
func stripLeadingVerbsPerSegment(s string) string {
	segments := splitByPipeRe.Split(s, -1)
	out := make([]string, 0, len(segments))
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		seg = splitByVerbAtStartRe.ReplaceAllString(seg, "")
		seg = splitByVerbAtEndRe.ReplaceAllString(seg, "")
		seg = strings.TrimSpace(seg)
		if seg != "" {
			out = append(out, seg)
		}
	}
	return strings.Join(out, " | ")
}
