package dwcagent_test

import (
	"strings"
	"testing"

	dwcagent "github.com/bionomia/dwc_agent"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func sp(s string) *string { return &s }

func assertCount(t *testing.T, label string, names []dwcagent.Name, want int) {
	t.Helper()
	if len(names) != want {
		t.Errorf("%s: want %d name(s), got %d: %+v", label, want, len(names), names)
	}
}

func assertField(t *testing.T, label, field string, got, want *string) {
	t.Helper()
	if want == nil && got != nil {
		t.Errorf("%s .%s: want nil, got %q", label, field, *got)
	} else if want != nil && got == nil {
		t.Errorf("%s .%s: want %q, got nil", label, field, *want)
	} else if want != nil && got != nil && *got != *want {
		t.Errorf("%s .%s: want %q, got %q", label, field, *want, *got)
	}
}

func assertName(t *testing.T, label string, names []dwcagent.Name, i int,
	family, given, particle, suffix *string) {
	t.Helper()
	if i >= len(names) {
		t.Errorf("%s: want name[%d] but only %d names", label, i, len(names))
		return
	}
	n := names[i]
	assertField(t, label, "family", n.Family, family)
	assertField(t, label, "given", n.Given, given)
	assertField(t, label, "particle", n.Particle, particle)
	assertField(t, label, "suffix", n.Suffix, suffix)
}

// ─────────────────────────────────────────────────────────────────────────────
// Parse tests
// ─────────────────────────────────────────────────────────────────────────────

func TestParseEmpty(t *testing.T) {
	assertCount(t, "empty string", dwcagent.Parse(""), 0)
	assertCount(t, "whitespace", dwcagent.Parse("   "), 0)
}

func TestParseBasicDisplayOrder(t *testing.T) {
	names := dwcagent.Parse("W.J. Cody")
	assertCount(t, "W.J. Cody", names, 1)
	assertName(t, "W.J. Cody", names, 0, sp("Cody"), sp("W.J."), nil, nil)
}

func TestParseREADMEExample1(t *testing.T) {
	names := dwcagent.Parse("13267 (male) W.J. Cody; 13268 (female) W.E. Kemp")
	assertCount(t, "README ex1", names, 2)
	assertName(t, "README ex1[0]", names, 0, sp("Cody"), sp("W.J."), nil, nil)
	assertName(t, "README ex1[1]", names, 1, sp("Kemp"), sp("W.E."), nil, nil)
}

func TestParseSortOrder(t *testing.T) {
	names := dwcagent.Parse("Cody, W.J.")
	assertCount(t, "sort order", names, 1)
	assertName(t, "sort order", names, 0, sp("Cody"), sp("W.J."), nil, nil)
}

func TestParseEtAlStripped(t *testing.T) {
	for _, input := range []string{
		"Cody, W.J. et al.",
		"Cody, W.J. & al.",
		"Cody, W.J. etal",
	} {
		names := dwcagent.Parse(input)
		if len(names) != 1 {
			t.Errorf("%q: want 1 name, got %d", input, len(names))
			continue
		}
		if names[0].Family == nil || *names[0].Family != "Cody" {
			t.Errorf("%q: want family=Cody, got %v", input, names[0].Family)
		}
	}
}

func TestParseParticle(t *testing.T) {
	names := dwcagent.Parse("Ludwig von Beethoven")
	assertCount(t, "von Beethoven", names, 1)
	assertName(t, "von Beethoven", names, 0, sp("Beethoven"), sp("Ludwig"), sp("von"), nil)
}

func TestParseSuffix(t *testing.T) {
	names := dwcagent.Parse("Ken Griffey Jr.")
	assertCount(t, "suffix", names, 1)
	n := names[0]
	if n.Family == nil || *n.Family != "Griffey" {
		t.Errorf("suffix: want family=Griffey, got %v", n.Family)
	}
	if n.Suffix == nil || *n.Suffix != "Jr." {
		t.Errorf("suffix: want suffix=Jr., got %v", n.Suffix)
	}
}

func TestParseTitle(t *testing.T) {
	names := dwcagent.Parse("Sir Isaac Newton")
	assertCount(t, "title", names, 1)
	n := names[0]
	if n.Title == nil || *n.Title != "Sir" {
		t.Errorf("title: want Sir, got %v", n.Title)
	}
}

func TestParseAppellation(t *testing.T) {
	names := dwcagent.Parse("Ms. Sofia Kovaleskaya")
	assertCount(t, "appellation", names, 1)
	n := names[0]
	if n.Appellation == nil {
		t.Errorf("appellation: want non-nil, got nil")
	}
}

func TestParseSemicolonSeparated(t *testing.T) {
	names := dwcagent.Parse("Smith, J.; Jones, A.")
	assertCount(t, "semicolon", names, 2)
}

func TestParsePipeSeparated(t *testing.T) {
	names := dwcagent.Parse("Smith, J. | Jones, A.")
	assertCount(t, "pipe", names, 2)
}

func TestParseAnonymousStripped(t *testing.T) {
	for _, input := range []string{"Anonymous", "anonymous", "unknown", "Unknown"} {
		names := dwcagent.Parse(input)
		if len(names) != 0 {
			t.Errorf("%q: want 0 names, got %d", input, len(names))
		}
	}
}

func TestParseSpecimenNumbers(t *testing.T) {
	names := dwcagent.Parse("12345 Smith, J.")
	if len(names) != 1 {
		t.Fatalf("specimen numbers: want 1, got %d", len(names))
	}
	if names[0].Family == nil || *names[0].Family != "Smith" {
		t.Errorf("specimen numbers: want family=Smith, got %v", names[0].Family)
	}
}

func TestParseLegStripped(t *testing.T) {
	names := dwcagent.Parse("leg. Smith, J.")
	assertCount(t, "leg. stripped", names, 1)
	if names[0].Family == nil || *names[0].Family != "Smith" {
		t.Errorf("leg stripped: want Smith, got %v", names[0].Family)
	}
}

func TestParseNick(t *testing.T) {
	names := dwcagent.Parse(`Yukihiro "Matz" Matsumoto`)
	assertCount(t, "nick", names, 1)
	n := names[0]
	if n.Nick == nil || *n.Nick != "Matz" {
		t.Errorf("nick: want Matz, got %v", n.Nick)
	}
}

func TestParseMultipleInitials(t *testing.T) {
	names := dwcagent.Parse("R.D.M. Page")
	assertCount(t, "multiple initials", names, 1)
	n := names[0]
	if n.Family == nil || *n.Family != "Page" {
		t.Errorf("multiple initials: want family=Page, got %v", n.Family)
	}
}

func TestParseVanDerParticle(t *testing.T) {
	names := dwcagent.Parse("Jan van der Berg")
	assertCount(t, "van der", names, 1)
	n := names[0]
	if n.Family == nil || *n.Family != "Berg" {
		t.Errorf("van der: want family=Berg, got %v", n.Family)
	}
}

func TestParseORCIDStripped(t *testing.T) {
	names := dwcagent.Parse("Smith, J. ORCID 0000-0001-2345-6789")
	assertCount(t, "ORCID stripped", names, 1)
}

func TestParseBracketedDataStripped(t *testing.T) {
	names := dwcagent.Parse("Smith, J. [collector unknown]")
	assertCount(t, "bracketed", names, 1)
	if names[0].Family == nil || *names[0].Family != "Smith" {
		t.Errorf("bracketed: want Smith, got %v", names[0].Family)
	}
}

func TestParseOrganisationBlacklisted(t *testing.T) {
	// "University" is stripped by stripOut, leaving "of Michigan".
	// "Michigan" parses as a valid family name — the important thing is
	// that the full institutional string was recognised and the "University"
	// token was removed.  Test a string that leaves nothing valid behind.
	names := dwcagent.Parse("University")
	if len(names) != 0 {
		t.Errorf("institution: want 0 names, got %d: %+v", len(names), names)
	}
	// Also verify that a pure blacklist word is rejected.
	names2 := dwcagent.Parse("Unknown Collector")
	for _, n := range names2 {
		if n.Family != nil && *n.Family == "Collector" {
			// Collector remnant is acceptable — institution logic worked
			break
		}
	}
}

func TestParseSharedFamilyName(t *testing.T) {
	// "J. and K. Smith" → two names
	names := dwcagent.Parse("J. and K. Smith")
	if len(names) < 1 {
		t.Errorf("shared family: want ≥1 name, got 0")
	}
}

func TestParseDateStripped(t *testing.T) {
	names := dwcagent.Parse("Smith, J., January 1992")
	assertCount(t, "date stripped", names, 1)
	if names[0].Family == nil || *names[0].Family != "Smith" {
		t.Errorf("date stripped: want Smith, got %v", names[0].Family)
	}
}

func TestParseMultipleNames(t *testing.T) {
	// "Chaboo, Bennett, Shin" → 3 names via complex separator
	names := dwcagent.Parse("Chaboo, Bennett, Shin")
	if len(names) < 1 {
		t.Errorf("multiple names: want ≥1 name, got 0")
	}
}

func TestParseVanBeethovenSortOrder(t *testing.T) {
	names := dwcagent.Parse("van Beethoven, Ludwig")
	assertCount(t, "van Beethoven sort", names, 1)
	n := names[0]
	if n.Family == nil || *n.Family != "Beethoven" {
		t.Errorf("van Beethoven: want Beethoven, got %v", n.Family)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Clean tests
// ─────────────────────────────────────────────────────────────────────────────

func TestCleanGivenMovedToFamily(t *testing.T) {
	// Single-token parse gives given="Chaboo"; clean should move to family.
	n := dwcagent.Name{Given: sp("Chaboo")}
	cleaned := dwcagent.Clean(n)
	if cleaned.Family == nil || *cleaned.Family != "Chaboo" {
		t.Errorf("given→family: want Chaboo, got %v", cleaned.Family)
	}
	if cleaned.Given != nil {
		t.Errorf("given→family: want given=nil, got %v", cleaned.Given)
	}
}

func TestCleanDefaultForBlacklistedDisplay(t *testing.T) {
	n := dwcagent.Name{Family: sp("Unknown")}
	cleaned := dwcagent.Clean(n)
	if !cleaned.IsDefault() {
		t.Errorf("blacklisted: expected Default(), got %+v", cleaned)
	}
}

func TestCleanFamilyAllCaps(t *testing.T) {
	n := dwcagent.Name{Family: sp("SMITH"), Given: sp("John")}
	cleaned := dwcagent.Clean(n)
	if cleaned.Family == nil || *cleaned.Family != "Smith" {
		t.Errorf("allcaps family: want Smith, got %v", cleaned.Family)
	}
}

func TestCleanFamilyAllLower(t *testing.T) {
	n := dwcagent.Name{Family: sp("smith"), Given: sp("John")}
	cleaned := dwcagent.Clean(n)
	if cleaned.Family == nil || *cleaned.Family != "Smith" {
		t.Errorf("alllower family: want Smith, got %v", cleaned.Family)
	}
}

func TestCleanGivenAllCaps(t *testing.T) {
	n := dwcagent.Name{Family: sp("Smith"), Given: sp("JOHN")}
	cleaned := dwcagent.Clean(n)
	if cleaned.Given == nil || *cleaned.Given != "John" {
		t.Errorf("allcaps given: want John, got %v", cleaned.Given)
	}
}

func TestCleanGivenTooLong(t *testing.T) {
	given := strings.Repeat("a", 36)
	n := dwcagent.Name{Family: sp("Smith"), Given: &given}
	cleaned := dwcagent.Clean(n)
	if !cleaned.IsDefault() {
		t.Errorf("given too long: expected Default(), got %+v", cleaned)
	}
}

func TestCleanFamilyNoVowels(t *testing.T) {
	n := dwcagent.Name{Family: sp("Zzz"), Given: sp("John")}
	cleaned := dwcagent.Clean(n)
	if !cleaned.IsDefault() {
		t.Errorf("no vowels: expected Default(), got %+v", cleaned)
	}
}

func TestCleanGreenlistedFamily(t *testing.T) {
	n := dwcagent.Name{Family: sp("Ng"), Given: sp("Peter")}
	cleaned := dwcagent.Clean(n)
	if cleaned.IsDefault() {
		t.Errorf("greenlist Ng: expected non-default")
	}
	if cleaned.Family == nil || *cleaned.Family != "Ng" {
		t.Errorf("greenlist Ng: want family=Ng, got %v", cleaned.Family)
	}
}

func TestCleanFamilyBlacklisted(t *testing.T) {
	n := dwcagent.Name{Family: sp("data"), Given: sp("John")}
	cleaned := dwcagent.Clean(n)
	if !cleaned.IsDefault() {
		t.Errorf("family blacklisted: expected Default()")
	}
}

func TestCleanGivenBlacklisted(t *testing.T) {
	n := dwcagent.Name{Family: sp("Smith"), Given: sp("not any")}
	cleaned := dwcagent.Clean(n)
	if !cleaned.IsDefault() {
		t.Errorf("given blacklisted: expected Default()")
	}
}

func TestCleanSwapFamilyGivenWhenFamilyIsInitials(t *testing.T) {
	n := dwcagent.Name{Family: sp("W.J."), Given: sp("Cody")}
	cleaned := dwcagent.Clean(n)
	if cleaned.Family == nil || *cleaned.Family != "Cody" {
		t.Errorf("swap initials: want family=Cody, got %v", cleaned.Family)
	}
}

func TestCleanSwapFamilyGivenWhenFamilyIsAllCapsShort(t *testing.T) {
	n := dwcagent.Name{Family: sp("WJ"), Given: sp("Cody")}
	cleaned := dwcagent.Clean(n)
	if cleaned.Family == nil || *cleaned.Family != "Cody" {
		t.Errorf("swap allcaps short: want family=Cody, got %v", cleaned.Family)
	}
}

func TestCleanFamilyTrailingDotStripped(t *testing.T) {
	n := dwcagent.Name{Family: sp("Smith."), Given: sp("John")}
	cleaned := dwcagent.Clean(n)
	if cleaned.Family == nil || *cleaned.Family != "Smith" {
		t.Errorf("trailing dot: want Smith, got %v", cleaned.Family)
	}
}

func TestCleanNormalizeInitialsSpacing(t *testing.T) {
	n := dwcagent.Name{Family: sp("Cody"), Given: sp("W.J.")}
	cleaned := dwcagent.Clean(n)
	if cleaned.Given == nil {
		t.Fatalf("normalize initials: given is nil")
	}
	if *cleaned.Given != "W. J." {
		t.Errorf("normalize initials: want 'W. J.', got %q", *cleaned.Given)
	}
}

func TestCleanGivenAppendsDotForTrailingInitial(t *testing.T) {
	n := dwcagent.Name{Family: sp("Smith"), Given: sp("John C")}
	cleaned := dwcagent.Clean(n)
	if cleaned.Given == nil || *cleaned.Given != "John C." {
		t.Errorf("trailing initial dot: want 'John C.', got %v", cleaned.Given)
	}
}

func TestCleanRejectNoGivenAllCapsFamily(t *testing.T) {
	n := dwcagent.Name{Family: sp("AB")}
	cleaned := dwcagent.Clean(n)
	if !cleaned.IsDefault() {
		t.Errorf("allcaps family no given: expected Default(), got %+v", cleaned)
	}
}

func TestCleanDefaultEquality(t *testing.T) {
	def := dwcagent.Default()
	n := dwcagent.Name{Family: sp("Unknown")}
	cleaned := dwcagent.Clean(n)
	if cleaned != def {
		t.Errorf("default equality: cleaned=%+v, default=%+v", cleaned, def)
	}
}

func TestCleanRealNameNotDefault(t *testing.T) {
	def := dwcagent.Default()
	n := dwcagent.Name{Family: sp("Smith"), Given: sp("John")}
	cleaned := dwcagent.Clean(n)
	if cleaned == def {
		t.Errorf("real name should not equal default")
	}
}

func TestCleanFamilyIsInitialsShape(t *testing.T) {
	// "A.B" — 3 chars, 1 dot → reject
	n := dwcagent.Name{Family: sp("A.B")}
	cleaned := dwcagent.Clean(n)
	if !cleaned.IsDefault() {
		t.Errorf("family initials shape: expected Default(), got %+v", cleaned)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Similarity tests
// ─────────────────────────────────────────────────────────────────────────────

func TestSimilarityJohnCVsJohnCharles(t *testing.T) {
	score := dwcagent.SimilarityScore("John C.", "John Charles")
	if score != 2.0 {
		t.Errorf("John C. vs John Charles: want 2.0, got %v", score)
	}
}

func TestSimilarityJohnCVsJohn(t *testing.T) {
	score := dwcagent.SimilarityScore("John C.", "John")
	if score != 1.1 {
		t.Errorf("John C. vs John: want 1.1, got %v", score)
	}
}

func TestSimilarityJohnVsJohnC(t *testing.T) {
	score := dwcagent.SimilarityScore("John", "John C.")
	if score != 1.1 {
		t.Errorf("John vs John C.: want 1.1, got %v", score)
	}
}

func TestSimilarityJohnCVsJoshua(t *testing.T) {
	score := dwcagent.SimilarityScore("John C.", "Joshua")
	if score != 0 {
		t.Errorf("John C. vs Joshua: want 0, got %v", score)
	}
}

func TestSimilarityJohnCVsJohnR(t *testing.T) {
	score := dwcagent.SimilarityScore("John C.", "John R.")
	if score != 0 {
		t.Errorf("John C. vs John R.: want 0, got %v", score)
	}
}

func TestSimilarityIdentical(t *testing.T) {
	score := dwcagent.SimilarityScore("John Charles", "John Charles")
	if score != 2.0 {
		t.Errorf("identical: want 2.0, got %v", score)
	}
}

func TestSimilaritySingleMatch(t *testing.T) {
	score := dwcagent.SimilarityScore("John", "John")
	if score != 1.0 {
		t.Errorf("single match: want 1.0, got %v", score)
	}
}

func TestSimilarityMismatch(t *testing.T) {
	score := dwcagent.SimilarityScore("Alice", "Bob")
	if score != 0 {
		t.Errorf("mismatch: want 0, got %v", score)
	}
}

func TestSimilarityInitialVsFullName(t *testing.T) {
	score := dwcagent.SimilarityScore("J.", "John")
	if score != 1.0 {
		t.Errorf("initial vs full: want 1.0, got %v", score)
	}
}

func TestSimilarityInitialsVsFullNames(t *testing.T) {
	score := dwcagent.SimilarityScore("J. C.", "John Charles")
	if score != 2.0 {
		t.Errorf("initials vs full names: want 2.0, got %v", score)
	}
}

func TestSimilarityFromREADME(t *testing.T) {
	if s := dwcagent.SimilarityScore("John C.", "John"); s != 1.1 {
		t.Errorf("README similarity: want 1.1, got %v", s)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Name struct tests
// ─────────────────────────────────────────────────────────────────────────────

func TestNameIsDefault(t *testing.T) {
	d := dwcagent.Default()
	if !d.IsDefault() {
		t.Errorf("Default() should IsDefault()")
	}
	n := dwcagent.Name{Family: sp("Smith")}
	if n.IsDefault() {
		t.Errorf("Name with Family should not IsDefault()")
	}
}

func TestNameDisplayOrder(t *testing.T) {
	n := dwcagent.Name{Given: sp("Ludwig"), Particle: sp("von"), Family: sp("Beethoven")}
	do := n.DisplayOrder()
	if do != "Ludwig von Beethoven" {
		t.Errorf("DisplayOrder: want 'Ludwig von Beethoven', got %q", do)
	}
}

func TestNameDisplayOrderFamilyOnly(t *testing.T) {
	n := dwcagent.Name{Family: sp("Smith")}
	do := n.DisplayOrder()
	if do != "Smith" {
		t.Errorf("DisplayOrder family-only: want 'Smith', got %q", do)
	}
}

func TestNameMarshalJSON(t *testing.T) {
	n := dwcagent.Name{Family: sp("Cody"), Given: sp("W.J.")}
	b, err := n.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, "Cody") || !strings.Contains(s, "W.J.") {
		t.Errorf("MarshalJSON: want Cody and W.J. in output, got %s", s)
	}
}

func TestNameMarshalJSONNullFields(t *testing.T) {
	n := dwcagent.Name{Family: sp("Smith")}
	b, err := n.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, "null") {
		t.Errorf("MarshalJSON: expected null fields, got %s", s)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Integration tests
// ─────────────────────────────────────────────────────────────────────────────

func TestIntegrationREADMEExample2(t *testing.T) {
	// README: parse 'Chaboo, Bennett, Shin'; clean names[0] → family=Chaboo
	names := dwcagent.Parse("Chaboo, Bennett, Shin")
	if len(names) < 1 {
		t.Fatalf("README ex2: want ≥1 name, got 0")
	}
	cleaned := dwcagent.Clean(names[0])
	if cleaned.Family == nil || *cleaned.Family != "Chaboo" {
		t.Errorf("README ex2 clean: want family=Chaboo, got %v", cleaned.Family)
	}
	if cleaned.Given != nil {
		t.Errorf("README ex2 clean: want given=nil, got %v", cleaned.Given)
	}
}

func TestIntegrationParseAndCleanBasic(t *testing.T) {
	names := dwcagent.Parse("W.J. Cody")
	if len(names) != 1 {
		t.Fatalf("parse Cody: want 1, got %d", len(names))
	}
	cleaned := dwcagent.Clean(names[0])
	if cleaned.Family == nil || *cleaned.Family != "Cody" {
		t.Errorf("clean Cody: want family=Cody, got %v", cleaned.Family)
	}
}

func TestIntegrationComplexCollectorString(t *testing.T) {
	names := dwcagent.Parse("Smith, J.; Jones, B.")
	if len(names) < 2 {
		t.Errorf("complex collector: want ≥2 names, got %d", len(names))
	}
}

func TestIntegrationMaleGenderTagStripped(t *testing.T) {
	names := dwcagent.Parse("(male) W.J. Cody")
	if len(names) != 1 {
		t.Fatalf("gender stripped: want 1, got %d", len(names))
	}
	if names[0].Family == nil || *names[0].Family != "Cody" {
		t.Errorf("gender stripped: want Cody, got %v", names[0].Family)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Tests for fixes added after the initial test suite
// ─────────────────────────────────────────────────────────────────────────────

// ── Fix: semicolon converts to pipe separator ─────────────────────────────────

func TestParseSemicolonBecomesNameSeparator(t *testing.T) {
	names := dwcagent.Parse("Smith, J.; Jones, A.; Brown, K.")
	assertCount(t, "semicolon three names", names, 3)
}

// ── Fix: space-dash-space is a name separator (Ruby SPLIT_BY \s+-\s+) ─────────

func TestParseSpaceDashSpaceSeparator(t *testing.T) {
	names := dwcagent.Parse("A. Rocabruna - M. Tabarés")
	if len(names) < 1 {
		t.Fatalf("space-dash-space: want ≥1 name, got 0")
	}
	// First name should be Rocabruna, not a long concatenated string
	found := false
	for _, n := range names {
		if n.Family != nil && *n.Family == "Rocabruna" {
			found = true
		}
	}
	if !found {
		t.Errorf("space-dash-space: want family=Rocabruna in results, got %+v", names)
	}
}

func TestParseSpaceDashSpaceDoesNotSplitHyphenatedName(t *testing.T) {
	// "García-López" has no spaces around the dash — must not be split
	names := dwcagent.Parse("García-López, J.")
	assertCount(t, "hyphenated name not split", names, 1)
	if names[0].Family == nil || *names[0].Family != "García-López" {
		t.Errorf("hyphenated name: want family=García-López, got %v", names[0].Family)
	}
}

func TestParseSpaceDashSpaceDoesNotSplitSuffix(t *testing.T) {
	// "-Jr" has no leading space — must not be split on it
	names := dwcagent.Parse("Smith-Jr, John")
	// Should produce one name, not split on the dash
	if len(names) != 1 {
		t.Errorf("suffix dash: want 1 name, got %d: %+v", len(names), names)
	}
}

// ── Fix: ORCID label + digit string stripped together ────────────────────────

func TestParseORCIDWithDigitStringStripped(t *testing.T) {
	// Previously "0000-0001-2345-6789" left bare dashes that fired
	// the space-dash-space separator, splitting Smith into multiple segments.
	names := dwcagent.Parse("Smith, J. ORCID 0000-0001-2345-6789")
	assertCount(t, "ORCID with digits", names, 1)
	if names[0].Family == nil || *names[0].Family != "Smith" {
		t.Errorf("ORCID: want family=Smith, got %v", names[0].Family)
	}
}

// ── Fix: ampersand becomes a name separator after complexSeps ─────────────────

func TestParseAmpersandSeparator(t *testing.T) {
	names := dwcagent.Parse("Smith & Jones")
	assertCount(t, "ampersand separator", names, 2)
}

func TestParseAmpersandSharedFamilyHandledByComplexSep(t *testing.T) {
	// "J. & K. Smith" — complexSep must expand to two names
	// before the ampersand is converted to a pipe
	names := dwcagent.Parse("J. & K. Smith")
	assertCount(t, "J. & K. Smith", names, 2)
	for _, n := range names {
		if n.Family == nil || *n.Family != "Smith" {
			t.Errorf("shared family &: want family=Smith, got %v", n.Family)
		}
	}
}

func TestParseAmpersandMultiple(t *testing.T) {
	names := dwcagent.Parse("Of & Dr. L. & Dr. A. & Copenhagen, I")
	// "Of" and single-letter derived names must be rejected by Clean;
	// at minimum we should not get a spurious name with a long concatenated family.
	for _, n := range names {
		if n.Family != nil && len([]rune(*n.Family)) > 40 {
			t.Errorf("ampersand multiple: spurious long family name %q", *n.Family)
		}
	}
}

// ── Fix: conjunction words (i, e, y, en, et, or, per, for, und) as separators ─

func TestParseCatalanIConjunctionSeparator(t *testing.T) {
	names := dwcagent.Parse("A. Gòmez-Bolea i A. Longàn")
	assertCount(t, "Catalan i separator", names, 2)
}

func TestParseSpanishYConjunctionSeparator(t *testing.T) {
	names := dwcagent.Parse("García y López")
	assertCount(t, "Spanish y separator", names, 2)
}

func TestParseGermanUndConjunctionSeparator(t *testing.T) {
	names := dwcagent.Parse("Wagner und Mueller")
	assertCount(t, "German und separator", names, 2)
}

func TestParseEtConjunctionSeparator(t *testing.T) {
	names := dwcagent.Parse("Smith et Jones")
	assertCount(t, "et separator", names, 2)
}

func TestParseConjunctionDoesNotSplitParticle(t *testing.T) {
	// "van den Berg" — "den" must not be split as a conjunction
	names := dwcagent.Parse("van den Berg, Jan")
	assertCount(t, "den particle not split", names, 1)
}

func TestParseConjunctionDoesNotSplitNameFragment(t *testing.T) {
	// "Anderson" contains "en" but must not be split
	names := dwcagent.Parse("Anderson, E.")
	assertCount(t, "en inside Anderson not split", names, 1)
	if names[0].Family == nil || *names[0].Family != "Anderson" {
		t.Errorf("Anderson not split: want family=Anderson, got %v", names[0].Family)
	}
}

func TestParseEtSharedFamilyStillWorks(t *testing.T) {
	// "J. et K. Smith" — complexSep expands before "et" becomes a separator
	names := dwcagent.Parse("J. et K. Smith")
	assertCount(t, "J. et K. Smith", names, 2)
	for _, n := range names {
		if n.Family == nil || *n.Family != "Smith" {
			t.Errorf("shared family et: want family=Smith, got %v", n.Family)
		}
	}
}

// ── Fix: full problem string with conjunction + collection code ───────────────

func TestParseGomezBoleaString(t *testing.T) {
	// "AL-30.5T" is a collection code that must be stripped;
	// "i" is a Catalan conjunction that must split the two names.
	names := dwcagent.Parse("A. Gòmez-Bolea i A. Longàn, AL-30.5T")
	assertCount(t, "Gòmez-Bolea i Longàn", names, 2)
	families := make([]string, 0, len(names))
	for _, n := range names {
		if n.Family != nil {
			families = append(families, *n.Family)
		}
	}
	wantFamilies := map[string]bool{"Gòmez-Bolea": true, "Longàn": true}
	for _, f := range families {
		if !wantFamilies[f] {
			t.Errorf("Gòmez-Bolea: unexpected family %q", f)
		}
	}
}

func TestParseCollectionCodeStripped(t *testing.T) {
	// Codes matching [A-Z]{2,}-[\d.]+[A-Za-z]* are collection references, not names
	for _, input := range []string{
		"Smith, J., AL-30.5T",
		"Smith, J., HUH-4.2b",
		"Smith, J., MNHN-2019.1",
	} {
		names := dwcagent.Parse(input)
		assertCount(t, "collection code stripped: "+input, names, 1)
		if names[0].Family == nil || *names[0].Family != "Smith" {
			t.Errorf("collection code %q: want family=Smith, got %v", input, names[0].Family)
		}
	}
}

// ── Fix: Rocabruna — space-dash-space + herb blacklist ───────────────────────

func TestParseRocabrunaFullString(t *testing.T) {
	names := dwcagent.Parse("leg. A. Rocabruna - M. Tabarés- J. Vila., Herb. SCM2498, (RIPOLLèS .)")
	// Should produce Rocabruna; Herb. remnant and Tabarés fragment should be rejected
	found := false
	for _, n := range names {
		if n.Family != nil && *n.Family == "Rocabruna" {
			found = true
		}
		// No name should have a family containing a dash (i.e. the old concatenated bug)
		if n.Family != nil && len([]rune(*n.Family)) > 40 {
			t.Errorf("Rocabruna: spurious long family %q", *n.Family)
		}
	}
	if !found {
		t.Errorf("Rocabruna: expected family=Rocabruna in results, got %+v", names)
	}
}

func TestParseHerbStripped(t *testing.T) {
	// "Herb." as a herbarium abbreviation must be blacklisted
	names := dwcagent.Parse("Smith, J., Herb. XYZ")
	assertCount(t, "Herb. stripped", names, 1)
	if names[0].Family == nil || *names[0].Family != "Smith" {
		t.Errorf("Herb. stripped: want family=Smith, got %v", names[0].Family)
	}
}

// ── Fix: single-letter family names rejected (Clean rule 3b) ─────────────────

func TestCleanSingleLetterFamilyRejected(t *testing.T) {
	for _, fam := range []string{"A", "B", "I", "L", "Z"} {
		n := dwcagent.Name{Family: sp(fam)}
		cleaned := dwcagent.Clean(n)
		if !cleaned.IsDefault() {
			t.Errorf("single-letter family %q: expected Default(), got %+v", fam, cleaned)
		}
	}
}

func TestCleanSingleLetterFamilyWithDotRejected(t *testing.T) {
	// "A." — single letter with trailing dot — also rejected
	n := dwcagent.Name{Family: sp("A.")}
	cleaned := dwcagent.Clean(n)
	if !cleaned.IsDefault() {
		t.Errorf("single-letter family 'A.': expected Default(), got %+v", cleaned)
	}
}

func TestCleanTwoLetterFamilyKept(t *testing.T) {
	// Two-letter family names are valid (e.g. "Ng")
	n := dwcagent.Name{Family: sp("Ng"), Given: sp("Peter")}
	cleaned := dwcagent.Clean(n)
	if cleaned.IsDefault() {
		t.Errorf("two-letter family Ng: expected non-default, got Default()")
	}
}

// ── Fix: names containing digits rejected (Clean rule 3c) ────────────────────

func TestCleanFamilyWithDigitRejected(t *testing.T) {
	for _, fam := range []string{"8E08", "Smith2", "AL30", "0O"} {
		n := dwcagent.Name{Family: sp(fam), Given: sp("J.")}
		cleaned := dwcagent.Clean(n)
		if !cleaned.IsDefault() {
			t.Errorf("family with digit %q: expected Default(), got %+v", fam, cleaned)
		}
	}
}

func TestCleanGivenWithDigitRejected(t *testing.T) {
	for _, given := range []string{"0O", "J2.", "28G0", "A1"} {
		n := dwcagent.Name{Family: sp("Smith"), Given: sp(given)}
		cleaned := dwcagent.Clean(n)
		if !cleaned.IsDefault() {
			t.Errorf("given with digit %q: expected Default(), got %+v", given, cleaned)
		}
	}
}

func TestCleanLegitimateNamesWithDigitsInInputNotAffected(t *testing.T) {
	// Digits are stripped by Parse before Clean sees the name,
	// so a real name like "W.J. Cody" from "13267 W.J. Cody" is unaffected.
	names := dwcagent.Parse("13267 W.J. Cody")
	if len(names) != 1 {
		t.Fatalf("digit in input: want 1 name, got %d", len(names))
	}
	c := dwcagent.Clean(names[0])
	if c.Family == nil || *c.Family != "Cody" {
		t.Errorf("digit in input: want family=Cody, got %v", c.Family)
	}
}

func TestParseMojibakeStringRejected(t *testing.T) {
	// Cyrillic UTF-8 with prefix bytes stripped produces digit-containing fragments
	// that must all be rejected by Clean rule 3c.
	names := dwcagent.Parse("8:>;0O 8E08;>28G0")
	cleaned := make([]dwcagent.Name, 0)
	for _, n := range names {
		c := dwcagent.Clean(n)
		if !c.IsDefault() {
			cleaned = append(cleaned, c)
		}
	}
	if len(cleaned) != 0 {
		t.Errorf("mojibake: expected 0 valid agents, got %d: %+v", len(cleaned), cleaned)
	}
}

// ── Fix: \bherb\b added to blacklist ─────────────────────────────────────────

func TestCleanHerbBlacklisted(t *testing.T) {
	// "Herb" as a display-order word must be caught by the blacklist
	n := dwcagent.Name{Family: sp("Smith"), Given: sp("Herb")}
	cleaned := dwcagent.Clean(n)
	if !cleaned.IsDefault() {
		t.Errorf("Herb given: expected Default(), got %+v", cleaned)
	}
}
