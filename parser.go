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
	//     Any "&" not consumed by complexSeparators is a simple name separator,
	//     mirroring Ruby's Namae which splits on "&" as part of its SPLIT_BY.
	//     Must run AFTER processComplexSeps so shared-family patterns like
	//     "J. & K. Smith" are expanded correctly before "&" becomes a pipe.
	s = strings.ReplaceAll(s, " & ", " | ")

	// 4c. Split on conjunction words — mirrors Ruby's SPLIT_BY which includes
	//     \b(con|e|y|i|en|et|or|per|for|und|och)\b as name separators.
	//     "i" splits Catalan, "och" splits Swedish, etc.
	//     Must run AFTER processComplexSeps so shared-family patterns like
	//     "J. et K. Smith" are handled before "et" becomes a pipe.
	s = conjunctionSepRe.ReplaceAllString(s, " | ")

	// 4d. Split on role/verb phrases — mirrors Ruby's SPLIT_BY verb entries:
	//     "det.", "identified by", "confirmed by", "ex.", "verified", etc.
	//     These phrases introduce a new agent in a collector chain.
	//     Must run AFTER processComplexSeps for the same reason as 4c.
	s = splitByVerbRe.ReplaceAllString(s, " | ")

	// 4e. Split on remaining punctuation separators — mirrors Ruby's SPLIT_BY
	//     [–|ǀ∣｜│&+\/;:] for the cases not yet handled:
	//     en-dash (–), colon (:), and the Catalan "a." separator.
	s = splitByPunctuationRe.ReplaceAllString(s, " | ")

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
