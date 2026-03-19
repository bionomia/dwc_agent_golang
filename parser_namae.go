package dwcagent

import (
	"regexp"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
// Namae-compatible parser
//
// Implements the BibTeX name-parsing rules used by the Ruby Namae gem:
//   Sort-order:    "Family, Given"
//   Display-order: "Given Family"
// ─────────────────────────────────────────────────────────────────────────────

var (
	suffixTokenRe = regexp.MustCompile(`(?i)^(jr\.?|sr\.?|esq\.?|[IVX]{2,}\.?)$`)

	titleTokenRe = regexp.MustCompile(
		`(?i)^(sir|count(?:ess)?|colonel|gen\.|adm\.|col\.|maj\.|cmdr\.|lt\.|sgt\.|cpl\.|pvt\.|` +
			`prof\.?|dr\.?|dra\.|md\.?|ph\.?d\.?|rev\.?|mme\.?|abb[eé]\.?|ptre\.?|bro\.?|esq\.?|` +
			`doct(?:eu|o)r|father|cantor|vicar|p[eè]re|pastor|profa\.?|profª|rabbi|reverend|soeur|sister|professor)$`)

	appellationTokenRe = regexp.MustCompile(`(?i)^(mrs?\.?|ms\.?|miss|fr\.?|hr\.?|herr|frau)$`)

	particleSet = map[string]bool{
		"van": true, "von": true, "de": true, "del": true, "der": true,
		"di": true, "du": true, "da": true, "do": true, "dos": true,
		"el": true, "la": true, "le": true, "les": true, "des": true,
		"ap": true, "the": true, "of": true,
	}

	twoTokenParticles = map[string]bool{
		"van de": true, "van der": true, "von der": true, "van den": true,
	}
)

// parseNames splits preprocessed input on "|" and parses each segment.
func parseNames(preprocessed string) []Name {
	segments := splitByPipeRe.Split(preprocessed, -1)
	var results []Name
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		seg = residualTerminatorsRe.ReplaceAllString(seg, "")
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		n := parseOne(seg)
		if !n.IsDefault() {
			results = append(results, n)
		}
	}
	return results
}

// parseOne parses a single name string.
func parseOne(s string) Name {
	if commaIdx := strings.Index(s, ","); commaIdx > 0 {
		return parseSortOrder(s, commaIdx)
	}
	return parseDisplayOrder(s)
}

// parseSortOrder handles "Family, Given [Suffix]"
func parseSortOrder(s string, commaIdx int) Name {
	familyPart := strings.TrimSpace(s[:commaIdx])
	restPart := strings.TrimSpace(s[commaIdx+1:])

	n := Name{}

	restTokens := tokenise(restPart)
	restTokens, suffix := extractSuffix(restTokens)
	restTokens, title := extractTitle(restTokens)
	restTokens, appellation := extractAppellation(restTokens)

	if suffix != "" {
		n.Suffix = strPtr(suffix)
	}
	if title != "" {
		n.Title = strPtr(title)
	}
	if appellation != "" {
		n.Appellation = strPtr(appellation)
	}

	familyTokens := tokenise(familyPart)
	familyTokens, particle := extractLeadingParticle(familyTokens)
	family := strings.Join(familyTokens, " ")

	if family != "" {
		n.Family = strPtr(family)
	}
	if particle != "" {
		n.Particle = strPtr(particle)
	}

	given := strings.Join(restTokens, " ")
	if given != "" {
		n.Given = strPtr(given)
	}
	return n
}

// parseDisplayOrder handles "Given [Particle] Family [Suffix]"
func parseDisplayOrder(s string) Name {
	tokens := tokenise(s)
	if len(tokens) == 0 {
		return Default()
	}

	n := Name{}
	tokens, suffix := extractSuffix(tokens)
	tokens, title := extractTitle(tokens)
	tokens, appellation := extractAppellation(tokens)

	if suffix != "" {
		n.Suffix = strPtr(suffix)
	}
	if title != "" {
		n.Title = strPtr(title)
	}
	if appellation != "" {
		n.Appellation = strPtr(appellation)
	}

	if len(tokens) == 0 {
		return n
	}

	tokens, nick := extractNick(tokens)
	if nick != "" {
		n.Nick = strPtr(nick)
	}

	if len(tokens) == 0 {
		return n
	}

	// Single token → treat as given (cleaner will promote to family)
	if len(tokens) == 1 {
		n.Given = strPtr(tokens[0])
		return n
	}

	// Check for two-token particle second-to-last
	if len(tokens) >= 3 {
		maybeTwo := strings.ToLower(tokens[len(tokens)-3] + " " + tokens[len(tokens)-2])
		if twoTokenParticles[maybeTwo] {
			n.Given = strPtr(strings.Join(tokens[:len(tokens)-3], " "))
			n.Particle = strPtr(tokens[len(tokens)-3] + " " + tokens[len(tokens)-2])
			n.Family = strPtr(tokens[len(tokens)-1])
			return n
		}
	}

	// Check for single particle second-to-last
	if len(tokens) >= 2 {
		maybePart := strings.ToLower(tokens[len(tokens)-2])
		if particleSet[maybePart] {
			given := strings.Join(tokens[:len(tokens)-2], " ")
			if given != "" {
				n.Given = strPtr(given)
			}
			n.Particle = strPtr(tokens[len(tokens)-2])
			n.Family = strPtr(tokens[len(tokens)-1])
			return n
		}
	}

	// Default: last token = family, rest = given
	n.Family = strPtr(tokens[len(tokens)-1])
	given := strings.Join(tokens[:len(tokens)-1], " ")
	if given != "" {
		n.Given = strPtr(given)
	}
	return n
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func tokenise(s string) []string {
	s = dotThenWordRe.ReplaceAllString(s, "$1 $2")
	return strings.Fields(s)
}

func extractSuffix(tokens []string) ([]string, string) {
	if len(tokens) == 0 {
		return tokens, ""
	}
	last := tokens[len(tokens)-1]
	if suffixTokenRe.MatchString(last) {
		return tokens[:len(tokens)-1], last
	}
	return tokens, ""
}

func extractTitle(tokens []string) ([]string, string) {
	var titles []string
	for len(tokens) > 0 && titleTokenRe.MatchString(tokens[0]) {
		titles = append(titles, tokens[0])
		tokens = tokens[1:]
	}
	return tokens, strings.Join(titles, " ")
}

func extractAppellation(tokens []string) ([]string, string) {
	if len(tokens) > 0 && appellationTokenRe.MatchString(tokens[0]) {
		return tokens[1:], tokens[0]
	}
	return tokens, ""
}

func extractNick(tokens []string) ([]string, string) {
	for i, t := range tokens {
		if len(t) >= 2 && t[0] == '"' && t[len(t)-1] == '"' {
			nick := t[1 : len(t)-1]
			result := make([]string, 0, len(tokens)-1)
			result = append(result, tokens[:i]...)
			result = append(result, tokens[i+1:]...)
			return result, nick
		}
	}
	return tokens, ""
}

func extractLeadingParticle(tokens []string) ([]string, string) {
	if len(tokens) >= 2 {
		maybeTwo := strings.ToLower(tokens[0] + " " + tokens[1])
		if twoTokenParticles[maybeTwo] {
			return tokens[2:], tokens[0] + " " + tokens[1]
		}
	}
	if len(tokens) >= 1 && particleSet[strings.ToLower(tokens[0])] {
		return tokens[1:], tokens[0]
	}
	return tokens, ""
}

// normalizeInitials expands "W.J." to "W. J." in the Given field.
func normalizeInitials(n *Name) {
	if n.Given == nil {
		return
	}
	g := *n.Given
	re := regexp.MustCompile(`([A-Za-z]\.)([A-Za-z]\.)`)
	for re.MatchString(g) {
		g = re.ReplaceAllString(g, "$1 $2")
	}
	n.Given = strPtr(g)
}
