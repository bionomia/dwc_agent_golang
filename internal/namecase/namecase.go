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

func titleWord(word string) string {
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

	// Mac/Mc prefix: "mcdonald" → "McDonald", "macintosh" → "MacIntosh"
	if strings.HasPrefix(lower, "mac") && len(lower) > 3 {
		return "Mac" + capitalize(lower[3:])
	}
	if strings.HasPrefix(lower, "mc") && len(lower) > 2 {
		return "Mc" + capitalize(lower[2:])
	}

	return capitalize(word)
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
