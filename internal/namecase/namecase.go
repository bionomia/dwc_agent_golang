// Package namecase provides title-casing for personal names, mirroring the
// Ruby namecase gem behaviour used by dwc_agent.
package namecase

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// NameCase title-cases a personal name string applying common rules for
// prefixes (Mc/Mac), particles, apostrophes, and hyphens.
// It preserves existing internal capitalisation for tokens that contain a
// mix of upper and lower case letters (e.g. "McTaggart-Cowan" stays intact).
func NameCase(s string) string {
	if s == "" {
		return s
	}
	words := strings.Fields(s)
	out := make([]string, 0, len(words))
	for _, word := range words {
		out = append(out, titleWord(word))
	}
	return strings.Join(out, " ")
}

// titleWord applies namecase rules to a single space-free word.
func titleWord(word string) string {
	if word == "" {
		return word
	}

	// Dotted initials like "R.T." or "W.J." must be preserved exactly as-is —
	// they are already correctly formatted and must not be lowercased.
	if isDottedInitials(word) {
		return word
	}

	// If the word already has mixed case AND the mixed case looks intentional
	// (i.e. not a trailing-caps typo like "SmitH"), preserve it entirely.
	// "MacQuarrie" → preserve. "McTaggart-Cowan" → preserve.
	if isMixedCase(word) && !hasTrailingCaps(word) {
		return word
	}

	// Handle hyphenated names: "smith-jones" → "Smith-Jones"
	if strings.Contains(word, "-") {
		parts := strings.Split(word, "-")
		for i, p := range parts {
			parts[i] = titleWord(p)
		}
		return strings.Join(parts, "-")
	}

	// Handle apostrophe: "o'brien" → "O'Brien"
	if idx := strings.Index(word, "'"); idx > 0 {
		before := word[:idx+1]
		after := word[idx+1:]
		if after != "" {
			return capitalize(before) + capitalize(after)
		}
		return capitalize(before)
	}

	lower := strings.ToLower(word)

	// Mac/Mc prefix — only apply when the input was all-caps (already handled
	// mixed-case above).  "MCDONALD" → "McDonald", "MACINTOSH" → "MacIntosh".
	if strings.HasPrefix(lower, "mac") && len(lower) > 4 {
		return "Mac" + capitalize(lower[3:])
	}
	if strings.HasPrefix(lower, "mc") && len(lower) > 2 {
		return "Mc" + capitalize(lower[2:])
	}

	return capitalize(word)
}

// isDottedInitials reports whether s looks like a run of dotted initials,
// e.g. "R.T.", "W.J.K.", "A.B.C." These must be preserved as-is.
func isDottedInitials(s string) bool {
	if len(s) < 2 {
		return false
	}
	// Must match: one or more (Letter.)+  where Letter is [A-Za-z]
	for i, r := range s {
		if i%2 == 0 { // even positions: letters
			if !unicode.IsLetter(r) {
				return false
			}
		} else { // odd positions: dots
			if r != '.' {
				return false
			}
		}
	}
	// Must end with a dot
	return s[len(s)-1] == '.'
}

// isMixedCase reports whether s contains both uppercase and lowercase letters.
func isMixedCase(s string) bool {
	hasUpper := false
	hasLower := false
	for _, r := range s {
		if unicode.IsUpper(r) {
			hasUpper = true
		} else if unicode.IsLower(r) {
			hasLower = true
		}
		if hasUpper && hasLower {
			return true
		}
	}
	return false
}

// hasTrailingCaps reports whether s ends with an uppercase letter preceded
// by a lowercase letter — e.g. "SmitH" has 'H' after 't'.
// This pattern signals a typo rather than intentional casing like "MacX".
func hasTrailingCaps(s string) bool {
	runes := []rune(s)
	if len(runes) < 2 {
		return false
	}
	last := runes[len(runes)-1]
	prev := runes[len(runes)-2]
	return unicode.IsUpper(last) && unicode.IsLower(prev)
}

// capitalize uppercases the first rune of s and lowercases the rest.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToUpper(r)) + strings.ToLower(s[size:])
}
