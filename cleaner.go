package dwcagent

import (
	"regexp"
	"strings"

	"github.com/bionomia/dwc_agent/internal/namecase"
)

// Clean post-processes a parsed Name to produce a well-formed Darwin Core name.
// Returns Default() when the name should be discarded.
// Mirrors DwcAgent::Cleaner#clean.
func Clean(n Name) Name {
	// 1. Reject all-initials names.
	if n.Family != nil {
		display := n.DisplayOrder()
		noSpace := strings.Join(strings.Fields(display), "")
		if noSpace == buildInitials(display) && noSpace != "" {
			return Default()
		}
	}

	// 2. Given blacklist.
	if n.Given != nil && givenBlacklist[strings.ToLower(*n.Given)] {
		return Default()
	}

	// 3. Family looks like initials: 3-char with exactly one dot.
	if n.Family != nil {
		fam := *n.Family
		if len([]rune(fam)) == 3 && strings.Count(fam, ".") == 1 {
			return Default()
		}
	}

	// 3b. Reject single-letter family names (with or without trailing dot).
	//     These arise from parsing artifacts like "Dr. A." -> family="A" or "A."
	//     and are never meaningful agent names.
	if n.Family != nil {
		fam := strings.TrimRight(*n.Family, ".")
		if len([]rune(fam)) == 1 {
			return Default()
		}
	}

	// 3c. Reject names where family or given contains a digit.
	//     Human names never contain digits. Tokens with digits are artefacts:
	//     specimen codes, mojibake (e.g. Cyrillic UTF-8 with prefix bytes
	//     stripped leaving printable ASCII fragments like "8E08" or "0O"),
	//     or accession numbers that survived the stripOut pass.
	hasDigit := regexp.MustCompile(`\d`)
	if n.Family != nil && hasDigit.MatchString(*n.Family) {
		return Default()
	}
	if n.Given != nil && hasDigit.MatchString(*n.Given) {
		return Default()
	}

	// 4. Length, space, and period limits.
	//    Both family and given must be < 40 characters.
	//    Family must have ≤ 2 spaces and ≤ 4 periods.
	//    Given must have ≤ 5 periods.
	if n.Family != nil {
		fam := *n.Family
		if len([]rune(fam)) >= 40 {
			return Default()
		}
		if strings.Count(fam, " ") > 2 {
			return Default()
		}
		if strings.Count(fam, ".") > 4 {
			return Default()
		}
	}
	if n.Given != nil {
		giv := *n.Given
		if len([]rune(giv)) >= 40 {
			return Default()
		}
		if strings.Count(giv, ".") > 5 {
			return Default()
		}
	}

	// 5. Given has ≥3 dots and matches suspicious pattern.
	if n.Given != nil {
		g := *n.Given
		if strings.Count(g, ".") >= 3 {
			if regexp.MustCompile(`\.\s*[a-zA-Z]{4,}\s+[a-zA-Z]{1,}\.`).MatchString(g) {
				return Default()
			}
		}
	}

	// 6. DisplayOrder matches blacklist.
	if blacklistRe.MatchString(n.DisplayOrder()) {
		return Default()
	}

	// 7. Strip trailing dot from family when len>3 and exactly one trailing dot.
	if n.Family != nil {
		fam := *n.Family
		if strings.Count(fam, ".") == 1 && strings.HasSuffix(fam, ".") && len([]rune(fam)) > 3 {
			n.Family = strPtr(strings.TrimSuffix(fam, "."))
		}
	}

	// 8. Family contains dots and letter-count ≤3 — swap given/family.
	if n.Given != nil && n.Family != nil {
		fam := *n.Family
		dotCount := strings.Count(fam, ".")
		letterCount := len([]rune(fam)) - dotCount
		if dotCount > 0 && letterCount <= 3 {
			given := *n.Given
			n.Family = strPtr(given)
			n.Given = strPtr(fam)
		}
	}

	// 9. Family ≤3 chars all-caps, given doesn't end in dot — swap.
	if n.Given != nil && n.Family != nil {
		fam := *n.Family
		given := *n.Given
		if len([]rune(fam)) <= 3 &&
			fam == strings.ToUpper(fam) &&
			!strings.HasSuffix(given, ".") {
			n.Family = strPtr(given)
			n.Given = strPtr(fam)
		}
	}

	// 10. No given, particle present, family matches "Word Initials" pattern.
	if n.Given == nil && n.Particle != nil && n.Family != nil {
		re := regexp.MustCompile(`^([A-Za-z]{3,})\s+((?:[A-Z]\.\s?){1,})$`)
		if m := re.FindStringSubmatch(*n.Family); m != nil {
			n.Family = strPtr(m[1])
			n.Given = strPtr(strings.TrimSpace(m[2]))
		}
	}

	// 11. Title-case all-upper or all-lower given (len≥4, no dots).
	if n.Given != nil {
		g := *n.Given
		if isUpperOrLower(g) && !strings.Contains(g, ".") && len([]rune(g)) >= 4 {
			n.Given = strPtr(namecase.NameCase(g))
		}
	}

	// 12. Given ends in lone capital → append dot.
	if n.Given != nil {
		g := *n.Given
		if trailingInitialRe.MatchString(g) {
			n.Given = strPtr(g + ".")
		}
	}

	// 13. Given contains "X." pattern AND has a real word → apply NameCase.
	// Skip pure-initial strings like "W.J." to avoid lowercasing "J." → "j."
	if n.Given != nil {
		g := *n.Given
		hasRealWord := regexp.MustCompile(`[A-Za-z]{2,}`).MatchString(g)
		if regexp.MustCompile(`[A-Za-z]\.`).MatchString(g) && hasRealWord {
			n.Given = strPtr(namecase.NameCase(g))
		}
	}

	// 14. Family blacklist (pre-normalization).
	if n.Family != nil && familyBlacklist[strings.ToLower(*n.Family)] {
		return Default()
	}

	// 15. No family but given present → move given to family.
	if n.Family == nil && n.Given != nil {
		g := *n.Given
		n.Family = strPtr(strings.TrimSuffix(g, "."))
		n.Given = nil
		// Re-check rule 3b: the promoted value may itself be a single letter
		// (e.g. given="A." → family="A"), which rule 3b could not catch earlier
		// because family was nil at that point.
		if n.Family != nil {
			fam := strings.TrimRight(*n.Family, ".")
			if len([]rune(fam)) == 1 {
				return Default()
			}
		}
	}

	// 16. Normalize initials spacing.
	normalizeInitials(&n)

	// 17. Collect trimmed copies of all fields.
	family := trimField(n.Family)
	given := trimField(n.Given)
	particle := trimField(n.Particle)
	appellation := trimField(n.Appellation)
	suffix := trimField(n.Suffix)
	title := trimField(n.Title)

	// 18. Given contains "X.Word" → insert space.
	if given != nil {
		g := *given
		if regexp.MustCompile(`[A-Z]\.[A-Za-z]{2,}`).MatchString(g) {
			g = strings.TrimSpace(strings.ReplaceAll(g, ".", ". "))
			given = &g
		}
	}

	// 19. Particle present but no given, and particle is not a known particle
	//     → capitalize and treat as given.
	if family != nil && given == nil && particle != nil {
		p := *particle
		if !particles[strings.ToLower(p)] {
			cap := capitalizeFirst(p)
			given = &cap
			particle = nil
		}
	}

	// 20. Particle contains dot but not "v" → discard particle.
	if particle != nil {
		p := *particle
		if strings.Contains(p, ".") && !strings.Contains(p, "v") {
			particle = nil
		}
	}

	// 21-pre. No given and family starts with 2 uppercase letters → reject.
	// Must run BEFORE NameCase normalization (rules 21/22) which would change "AB" → "Ab".
	if given == nil && family != nil {
		f := *family
		if regexp.MustCompile(`^[A-Z]{2}`).MatchString(f) {
			return Default()
		}
	}

	// 21. Title-case all-upper or all-lower family.
	if family != nil {
		f := *family
		if isUpperOrLower(f) {
			f = namecase.NameCase(f)
			family = &f
		}
	}

	// 22. Family ends in 1-3 all-caps letters → upper then NameCase.
	if family != nil {
		f := *family
		if regexp.MustCompile(`[A-Z]{1,3}$`).MatchString(f) {
			f = namecase.NameCase(strings.ToUpper(f))
			family = &f
		}
	}

	// 23. (moved to 21-pre above)

	// 24. Final family blacklist.
	if family != nil && familyBlacklist[strings.ToLower(*family)] {
		return Default()
	}

	// 25. Final given blacklist.
	if given != nil && givenBlacklist[strings.ToLower(*given)] {
		return Default()
	}

	// 26. Family has no vowels and is not greenlisted → reject.
	if family != nil {
		f := *family
		if !containsVowel(f) && !familyGreenlist[strings.ToLower(f)] {
			return Default()
		}
	}

	return Name{
		Title:            title,
		Appellation:      appellation,
		Given:            given,
		Particle:         particle,
		Family:           family,
		Suffix:           suffix,
		Nick:             n.Nick,
		DroppingParticle: n.DroppingParticle,
	}
}

// trimField returns nil if the string is empty/nil, or a pointer to the trimmed value.
func trimField(p *string) *string {
	if p == nil {
		return nil
	}
	s := strings.TrimSpace(*p)
	if s == "" {
		return nil
	}
	return &s
}

// buildInitials returns "X.Y." for each word in s (first letter + dot).
func buildInitials(s string) string {
	var b strings.Builder
	for _, word := range strings.Fields(s) {
		runes := []rune(word)
		if len(runes) > 0 {
			b.WriteRune(runes[0])
			b.WriteRune('.')
		}
	}
	return b.String()
}

// capitalizeFirst uppercases the first rune of s.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	re := regexp.MustCompile(`[a-z]\.`)
	s = re.ReplaceAllStringFunc(s, strings.ToUpper)
	runes := []rune(s)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}
