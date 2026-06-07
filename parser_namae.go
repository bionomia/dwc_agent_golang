package dwcagent

import (
	"regexp"
	"strings"
	"unicode"
)

// ─────────────────────────────────────────────────────────────────────────────
// Namae-compatible parser
// ─────────────────────────────────────────────────────────────────────────────

var (
	suffixTokenRe = regexp.MustCompile(
		`(?i)^(jr\.?|sr\.?|esq\.?|[IVX]{2,}\.?)$`)

	titleTokenRe = regexp.MustCompile(
		`(?i)^(sir|count(?:ess)?|colonel|gen\.|adm\.|col\.|maj\.|cmdr\.|lt\.|sgt\.|cpl\.|pvt\.|` +
			`prof\.?|dr\.?|dra\.|md\.?|ph\.?d\.?|rev\.?|mme\.?|abb[eé]\.?|ptre\.?|bro\.?|` +
			`doct(?:eu|o)r|father|cantor|vicar|p[eè]re|pastor|profa\.?|profª|rabbi|reverend|soeur|sister|professor|` +
			`esq\.?)$`)

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

	// allCapsInitialsRe matches 2–4 uppercase letters (no dots).
	// "FAH"→F.A.H., "JH"→J.H., "CJ"→C.J. Excludes 5+ char words like "CHRIS".
	allCapsInitialsRe = regexp.MustCompile(`^[A-Z]{2,4}$`)

	// singleCapRe matches exactly one bare uppercase letter used as an initial.
	singleCapRe = regexp.MustCompile(`^[A-Z]$`)

	// adjacentInitialRe collapses "X. Y." → "X.Y."
	adjacentInitialRe = regexp.MustCompile(`([A-Z]\.) ([A-Z]\.)`)

	// singleInitBetweenWordsRe detects "Jack E Smith" (middle bare initial).
	singleInitBetweenWordsRe = regexp.MustCompile(
		`^([A-Z][a-z]+)\s+([A-Z])\s+([A-Z][a-z]+)$`)

	// singleDottedInitRe matches a single dotted initial like "A." or "K.".
	// Used to detect a trailing initial that was expanded from a bare uppercase.
	singleDottedInitRe = regexp.MustCompile(`^[A-Z]\.$`)

	// dottedInitialsRe matches one or more dotted initials like "T.M.A." or "W.J.".
	dottedInitialsRe = regexp.MustCompile(`^(?:[A-Z]\.){1,}$`)

	// hyphenatedGivenRe detects "Hsuan-Ching" style — hyphenated multi-part
	// given name that got separated into two tokens by the pipe splitter.
	// This fires in parseDisplayOrder when a lone hyphenated token appears.
	hyphenatedWordRe = regexp.MustCompile(`^[A-Z][a-z]+-[A-Z][a-z]+$`)
)

// parseNames splits preprocessed input on "|" and parses each segment.
func parseNames(preprocessed string) []Name {
	trailingDotRe  := regexp.MustCompile(`\s+\.\s*$`)
	trailingDashRe := regexp.MustCompile(`\s*-+\s*$`)

	segments := splitByPipeRe.Split(preprocessed, -1)
	var results []Name
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		seg = residualTerminatorsRe.ReplaceAllString(seg, "")
		// Strip trailing isolated dot (left after bracket removal mid-string)
		// and trailing dashes (left after numeric-date removal)
		seg = trailingDotRe.ReplaceAllString(seg, "")
		seg = trailingDashRe.ReplaceAllString(seg, "")
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

// parseOne parses a single name segment.
func parseOne(s string) Name {
	if commaIdx := strings.Index(s, ","); commaIdx > 0 {
		return parseSortOrder(s, commaIdx)
	}
	return parseDisplayOrder(s)
}

// parseSortOrder handles "Family, Given [Suffix]" — the sort-order BibTeX form.
// The given field is everything after the comma and is always treated as-is
// (never re-parsed as display-order), since in "Tanner, C.A." the restPart
// "C.A." is the initials of the given name, not "C." given + "A." family.
func parseSortOrder(s string, commaIdx int) Name {
	familyPart := strings.TrimSpace(s[:commaIdx])
	restPart   := strings.TrimSpace(s[commaIdx+1:])

	n := Name{}

	restTokens := tokenise(restPart)
	restTokens, suffix      := extractSuffix(restTokens)
	restTokens, title       := extractTitle(restTokens)
	restTokens, appellation := extractAppellation(restTokens)

	if suffix != "" {
		if strings.EqualFold(strings.TrimSuffix(suffix, "."), "esq") {
			title = "Esq."
		} else {
			n.Suffix = strPtr(normaliseSuffix(suffix))
		}
	}
	if title != ""       { n.Title       = strPtr(title)       }
	if appellation != "" { n.Appellation = strPtr(appellation) }

	familyTokens := tokenise(familyPart)
	familyTokens, particle := extractLeadingParticle(familyTokens)
	family := strings.Join(familyTokens, " ")
	if family   != "" { n.Family   = strPtr(family)   }
	if particle != "" { n.Particle = strPtr(particle)  }

	// Treat the entire restPart as given — never try to split it further.
	// This correctly handles "C.A.", "W.J.K.", "N.", "E.", "Carolyn J.", etc.
	given := condenseInitials(strings.Join(restTokens, " "))
	if given != "" {
		n.Given = strPtr(given)
	}
	return n
}

// parseDisplayOrder handles "Given [Particle] Family [Suffix]" — display order.
func parseDisplayOrder(s string) Name {
	// Special case: "Jack E Smith" — Word BareInitial Word → given="Jack E.", family="Smith"
	if m := singleInitBetweenWordsRe.FindStringSubmatch(s); m != nil {
		n := Name{}
		n.Given  = strPtr(m[1] + " " + m[2] + ".")
		n.Family = strPtr(m[3])
		return n
	}

	tokens := tokenise(s)
	if len(tokens) == 0 {
		return Default()
	}

	// Expand all-caps initials and single bare-cap tokens.
	tokens = expandCapsInitials(tokens)

	n := Name{}
	tokens, suffix      := extractSuffix(tokens)
	tokens, title       := extractTitle(tokens)
	tokens, appellation := extractAppellation(tokens)

	if suffix != "" {
		if strings.EqualFold(strings.TrimSuffix(suffix, "."), "esq") {
			title = "Esq."
		} else {
			n.Suffix = strPtr(normaliseSuffix(suffix))
		}
	}
	if title != ""       { n.Title       = strPtr(title)       }
	if appellation != "" { n.Appellation = strPtr(appellation) }

	if len(tokens) == 0 {
		return n
	}

	tokens, nick := extractNick(tokens)
	if nick != "" { n.Nick = strPtr(nick) }

	if len(tokens) == 0 { return n }

	// Single token → treat as given (cleaner will promote to family if needed).
	if len(tokens) == 1 {
		n.Given = strPtr(tokens[0])
		return n
	}

	// Check for two-token particle second-to-last ("van der", "von der", etc.)
	if len(tokens) >= 3 {
		maybeTwo := strings.ToLower(tokens[len(tokens)-3] + " " + tokens[len(tokens)-2])
		if twoTokenParticles[maybeTwo] {
			givenRaw := strings.Join(tokens[:len(tokens)-3], " ")
			if givenRaw != "" { n.Given = strPtr(condenseInitials(givenRaw)) }
			n.Particle = strPtr(tokens[len(tokens)-3] + " " + tokens[len(tokens)-2])
			n.Family   = strPtr(tokens[len(tokens)-1])
			return n
		}
	}

	// Check for single particle second-to-last.
	if len(tokens) >= 2 {
		maybePart := strings.ToLower(tokens[len(tokens)-2])
		if particleSet[maybePart] {
			givenRaw := strings.Join(tokens[:len(tokens)-2], " ")
			if givenRaw != "" { n.Given = strPtr(condenseInitials(givenRaw)) }
			n.Particle = strPtr(tokens[len(tokens)-2])
			n.Family   = strPtr(tokens[len(tokens)-1])
			return n
		}
	}

	// Default: last token = family, rest = given.
	// Exception: if the last token is a dotted single initial (e.g. "A.", "K.") or a
	// run of dotted initials (e.g. "T.M.A." after TMA was expanded), and there is at
	// least one preceding non-initial word, that word is the family name and all the
	// initial tokens form the given. This handles:
	//   "Julius A"      → tokens ["Julius","A."]      → family=Julius, given=A.
	//   "Utteridge TMA" → tokens ["Utteridge","T.","M.","A."] → family=Utteridge, given=T.M.A.
	//   "Imin K"        → tokens ["Imin","K."]         → family=Imin, given=K.
	if len(tokens) >= 2 {
		last := tokens[len(tokens)-1]
		isTrailingInit := singleDottedInitRe.MatchString(last) || dottedInitialsRe.MatchString(last)
		if isTrailingInit {
			// Find the last non-initial token — that is the family name.
			lastNonInit := -1
			for i := len(tokens) - 2; i >= 0; i-- {
				if !singleDottedInitRe.MatchString(tokens[i]) && !dottedInitialsRe.MatchString(tokens[i]) {
					lastNonInit = i
					break
				}
			}
			if lastNonInit >= 0 {
				n.Family = strPtr(tokens[lastNonInit])
				// Given = all tokens except the family, condensed.
				givenTokens := make([]string, 0, len(tokens)-1)
				givenTokens = append(givenTokens, tokens[:lastNonInit]...)
				givenTokens = append(givenTokens, tokens[lastNonInit+1:]...)
				givenRaw := strings.Join(givenTokens, " ")
				if givenRaw != "" {
					n.Given = strPtr(condenseInitials(givenRaw))
				}
				return n
			}
		}
	}
	n.Family = strPtr(tokens[len(tokens)-1])
	givenRaw := strings.Join(tokens[:len(tokens)-1], " ")
	if givenRaw != "" {
		n.Given = strPtr(condenseInitials(givenRaw))
	}
	return n
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func tokenise(s string) []string {
	// Split compact "J.R.Smith" → "J.R. Smith" before splitting into tokens.
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

// expandCapsInitials expands run-together or spaced uppercase initial tokens.
//   "FAH"       → ["F.", "A.", "H."]
//   "JH"        → ["J.", "H."]
//   ["A","Y"]   → ["A.", "Y."]   (single bare caps in a multi-token list)
//   ["Julius","A"]      → ["Julius","A."]   (trailing initial after name word)
//   ["Utteridge","TMA"] → ["Utteridge","T.","M.","A."]  (trailing run after name word)
// Does NOT expand the last token when it looks like a genuine family name
// (5+ chars, mixed-case, or the only token with no preceding word).
func expandCapsInitials(tokens []string) []string {
	if len(tokens) == 0 {
		return tokens
	}

	// Count caps-style tokens (single bare cap or 2-4 all-caps).
	capsCount := 0
	for _, t := range tokens {
		if allCapsInitialsRe.MatchString(t) || singleCapRe.MatchString(t) {
			capsCount++
		}
	}
	allAreCaps := capsCount == len(tokens)

	// Detect whether the last token is a trailing initial: a bare uppercase (1-4 chars)
	// that follows at least one non-caps word. In "Julius A" or "Utteridge TMA" the
	// bare cap/caps-run is the given initial, not the family name.
	trailingInitial := false
	if len(tokens) >= 2 && !allAreCaps {
		last := tokens[len(tokens)-1]
		if singleCapRe.MatchString(last) || allCapsInitialsRe.MatchString(last) {
			// Check that at least one preceding token is a "real" word (not a caps token)
			for _, t := range tokens[:len(tokens)-1] {
				if !singleCapRe.MatchString(t) && !allCapsInitialsRe.MatchString(t) {
					trailingInitial = true
					break
				}
			}
		}
	}

	out := make([]string, 0, len(tokens)+4)
	for i, tok := range tokens {
		isLast := i == len(tokens)-1

		if singleCapRe.MatchString(tok) {
			// Single bare cap: always expand to dotted initial when not last,
			// or when all tokens are caps-style, or when it's a trailing initial.
			if !isLast || allAreCaps || trailingInitial {
				out = append(out, tok+".")
				continue
			}
		}
		if allCapsInitialsRe.MatchString(tok) {
			// 2–4 all-caps: expand when not last, or all are caps, or only 2 chars,
			// or when it's a trailing initial after a real word.
			if !isLast || allAreCaps || len(tok) <= 2 || trailingInitial {
				for _, r := range tok {
					out = append(out, string(r)+".")
				}
				continue
			}
		}
		out = append(out, tok)
	}
	return out
}

// condenseInitials collapses spaced initials into compact dotted form.
//   "B. P. J." → "B.P.J."
//   "R.K. A."  → "R.K.A."
//   "T. L."    → "T.L."
//   "A A"      → "A.A."  (single bare caps get a dot via expandCapsInitials first)
func condenseInitials(s string) string {
	if s == "" {
		return s
	}
	// Add dot to bare single capital letters that appear as isolated tokens.
	// "A A" → "A. A." so that adjacentInitialRe can then collapse them.
	s = regexp.MustCompile(`(^| )([A-Z])( |$)`).ReplaceAllStringFunc(s, func(m string) string {
		trimmed := strings.TrimSpace(m)
		if []rune(trimmed)[0] >= 'A' && []rune(trimmed)[0] <= 'Z' && len([]rune(trimmed)) == 1 {
			prefix, suffix2 := "", ""
			if strings.HasPrefix(m, " ") { prefix = " " }
			if strings.HasSuffix(m, " ") { suffix2 = " " }
			return prefix + trimmed + "." + suffix2
		}
		return m
	})
	// Collapse "X. Y." → "X.Y."
	for adjacentInitialRe.MatchString(s) {
		s = adjacentInitialRe.ReplaceAllString(s, "$1$2")
	}
	return strings.TrimSpace(s)
}

// normaliseSuffix ensures suffix has correct capitalisation and trailing dot.
func normaliseSuffix(s string) string {
	upper := strings.ToUpper(strings.TrimSuffix(s, "."))
	switch upper {
	case "JR":
		return "Jr."
	case "SR":
		return "Sr."
	}
	if regexp.MustCompile(`^[IVX]+$`).MatchString(upper) {
		return upper + "."
	}
	return s
}

// normalizeInitials is kept for backward compatibility with cleaner.go.
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

// isUpperRune reports whether r is an uppercase Unicode letter.
func isUpperRune(r rune) bool {
	return unicode.IsUpper(r)
}