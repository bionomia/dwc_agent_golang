package dwcagent

import (
	"regexp"
	"strings"
	"unicode"
)

var extraSpaceRe = regexp.MustCompile(`\s{2,}`)

// allCapsWordRe matches a run of 5+ uppercase ASCII letters forming a whole word.
// We use 5+ (not 2+) so that short all-caps tokens like "TMA" (concatenated initials),
// "BERG", "CODY" (4-letter family names) are left for expand() and nc() to handle.
// Only words of 5+ chars like "RIBAS", "BARBOSA" need early normalisation so that
// structure-detection regexes (which expect a lowercase letter after the first uppercase)
// work correctly regardless of input capitalisation.
var allCapsWordRe = regexp.MustCompile(`\b[A-Z]{5,}\b`)

// normaliseAllCaps title-cases every all-uppercase word in s so that subsequent
// parsing logic (which uses lowercase letters to detect family names) works
// identically whether the collector entered "O.S. RIBAS" or "O.S. Ribas".
// Mixed-case words (e.g. "McDonald") and dotted initials ("O.S.") are left alone.
func normaliseAllCaps(s string) string {
	return allCapsWordRe.ReplaceAllStringFunc(s, func(w string) string {
		runes := []rune(w)
		runes[0] = unicode.ToUpper(runes[0])
		for i := 1; i < len(runes); i++ {
			runes[i] = unicode.ToLower(runes[i])
		}
		return string(runes)
	})
}

// Parse cleanses the input string and returns a slice of parsed Names.
// Mirrors DwcAgent.parse in the Ruby gem.
func Parse(input string) []Name {
	if strings.TrimSpace(input) == "" {
		return []Name{}
	}

	s := input

	// 1. Strip noise tokens.
	s = stripOut(s)

	// 1b. Normalise all-caps words to title-case ("RIBAS" → "Ribas") so that
	//     name-structure detection (DOL, sort-order, etc.) is case-insensitive.
	//     Single uppercase letters and dotted initials are unaffected.
	s = normaliseAllCaps(s)

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
