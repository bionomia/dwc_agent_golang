package dwcagent

import (
	"regexp"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
// Character substitutions (CHAR_SUBS in Ruby). Order matters.
// ─────────────────────────────────────────────────────────────────────────────
var charSubsList = []struct{ from, to string }{
	{", ph.d.", " Ph.D."},
	{", Ph.D.", " Ph.D."},
	{", bro.", " Bro."},
	{", Jr.,", " Jr.;"},
	{", Jr.", " Jr."},
	{",Jr.", " Jr."},
	{", Sr.", " Sr."},
	{",Sr.", " Sr."},
	{" jr.,", " Jr.;"},
	{" jr,", " Jr.;"},
	{"-jr", " Jr."},
	{"-Jr", " Jr."},
	{"Dr.", "Dr. "},
	{"prof.", "Prof. "},
	{" .;", ". ;"},
	{", &", " &"},
	{";", " | "},
	{"\u201c", "'"},
	{"|", " | "},
	{"\u01c0", " | "},
	{"\u2223", " | "},
	{"\u2502", " | "},
	{"(", " "},
	{")", " "},
	{"?", ""},
	{"!", ""},
	{"=", ""},
	{"#", ""},
	{"/", " / "},
	{"&", " & "},
	{"*", ""},
	{">", ""},
	{"<", ""},
	{"{", ""},
	{"}", ""},
	{"@", ""},
	{"%", ""},
	{"\\", ""},
	{"\u00b4", "'"},
	{"+", " | "},
}

func applyCharSubs(s string) string {
	for _, sub := range charSubsList {
		s = strings.ReplaceAll(s, sub.from, sub.to)
	}
	return s
}

// ─────────────────────────────────────────────────────────────────────────────
// Strip-out patterns. Go RE2 does NOT support (?i:...) inline flags.
// All case-insensitive patterns use (?i) at the start of the pattern.
// ─────────────────────────────────────────────────────────────────────────────
var stripOutPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)acc\s?#`),
	regexp.MustCompile(`["'-]{2,}`),
	regexp.MustCompile(`-\.\s`),
	regexp.MustCompile(`(?i)[,;]?\s*(?:1st|2nd|3rd|[4-9]th)`),
	// ORCID digit sequences must be stripped FIRST, before any other digit or
	// dash pattern fires and breaks the 4×4 structure that the pattern relies on.
	// e.g. \d*[A-Za-z]*\d*-\d*$ would strip "-6789" from the end first, leaving
	// "0000-0001-2345-" which no longer matches \b\d{4}-\d{4}-\d{4}-\d{4}\b.
	regexp.MustCompile(`\b\d{4}-\d{4}-\d{4}-\d{4}\b`),
	// Collection/specimen codes like "AL-30.5T", "HUH-4.2b" — strip before
	// the decimal pattern fires so the trailing dash is not left as an orphan.
	regexp.MustCompile(`\b[A-Z]{2,}-[\d.]+[A-Za-z]*\b`),
	regexp.MustCompile(`[,]?\s*?\d+\.\d+`),
	regexp.MustCompile(`[,]?\s*\([#NnOo.\s0-9-]*[0-9a-z]+\)\s*$`),
	regexp.MustCompile(`[,]?\s+#[0-9a-z]+$`),
	regexp.MustCompile(`(?i)[,]?\s*#*\s+\d+[-/\s][A-Z\d]+-?\d*[A-Za-z]*$`),
	// Numeric dates "21-12-1971" BEFORE the trailing-dash strip so the full
	// date is removed as a unit rather than piece by piece.
	regexp.MustCompile(`\b\d{1,2}[-/.]\d{1,2}[-/.]\d{2,4}\b`),
	regexp.MustCompile(`\d*[A-Za-z]*\d*-\d*$`),
	regexp.MustCompile(`\b\d+\(?[[:alpha:]]\)?\b`),
	// Slash-date "20/Aug./1980" — must be before digit strips so digits not consumed first
	regexp.MustCompile(`(?i)\d+\s*/\s*(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sept?|Oct|Nov|Dec)\.?\s*/\s*\d+`),
	// Roman-numeral date "19.II.1902" — must be BEFORE \b\d{2,}\b so the surrounding
	// digits are not consumed first (leaving orphan ".II." which fools the name parser).
	regexp.MustCompile(`(?i)\d{1,4}[/.]?(?:i|ii|iii|iv|v|vi|vii|viii|ix|x|xi|xii)[/.]\d{1,4}`),
	regexp.MustCompile(`\b\d{2,}\b`),  // strip standalone numbers (specimen IDs etc.)
	regexp.MustCompile(`[,;\s]+(?:et\.?\s+al|&\s+al)l?\.?`),
	regexp.MustCompile(`(?i)\b[,;]?\s*etal\.?`),
	regexp.MustCompile(`(?i)\b[,;]?\s*et\.al\.?`),
	regexp.MustCompile(`\b\s+(?:bis|ter)(?:\b|$)`),
	regexp.MustCompile(`\bu\.\s*a\.`),
	regexp.MustCompile(`(?i)\b[,;]?\s*(?:and|&)?\s*(?:others|party)\s*\b`),
	regexp.MustCompile(`(?i)\b[,;]?\s*etc\.?`),
	regexp.MustCompile(`(?i)\b[,;]?\s*exp\.?\s*(?:\b|$)`),
	regexp.MustCompile(`(?i)\b[,;]?\s*aboard[^$]+`),
	regexp.MustCompile(`(?i)\b[,;]?\s+on\b`),
	regexp.MustCompile(`(?i)\bunknown\s+or\s+anonymous`),
	regexp.MustCompile(`(?i)\b[,;]?\s*unkn?own\b`),
	regexp.MustCompile(`(?i)\b[,;]?\s*n/a\b`),
	regexp.MustCompile(`(?i)\b[,;]?\s*ann?onymous\b`),
	regexp.MustCompile(`(?i)\b[,;]?\s*\(?(?:undetermined|indeterminable|dummy|interim|accession|ill(?:eg|is)ible|scripsit|presumably?)\)?\b`),
	regexp.MustCompile(`(?i)\b[,;]?\s*(?:importer|gift):?\b`),
	// "person string" as a complete phrase (also covers "Person String" alone)
	regexp.MustCompile(`(?i)\bperson\s*string\b`),
	regexp.MustCompile(`(?i)\b[,;]?\s*string\b`),
	regexp.MustCompile(`(?i)^colln?\.?\s+|\s*colln?\.?\s*$`),
	regexp.MustCompile(`(?i)^collection:?\s+|\s*collection\s*$`),
	regexp.MustCompile(`(?i)\b[,;]?\s*colls\.(?:\b|$)`),
	regexp.MustCompile(`(?i)contactid`),
	regexp.MustCompile(`(?i)^dupl[.,]+`),
	regexp.MustCompile(`(?i)\b[,;]?\s*stet[,!]?\s*\d*$`),
	// Date patterns: allow space OR dash/slash between day-number and month name
	// so "21 Dec. 1999", "21-12-1971", and "20/Aug./1980" are all caught.
	regexp.MustCompile(`(?i)[,;]?\s*\d+[-/.\s]?(?:\d+|Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sept?|Oct|Nov|Dec)\.?\s*[-/\s]?\d+`),
	// Also strip "20/Aug./1980" form where month is surrounded by slashes
	regexp.MustCompile(`(?i)\d+\s*/\s*(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sept?|Oct|Nov|Dec)\.?\s*/\s*\d+`),
	regexp.MustCompile(`(?i)\b[,;]?\s*(?:Jan|January|janvier)[.,;]?\s*\d+`),
	regexp.MustCompile(`(?i)\b[,;]?\s*(?:Feb|February|f(?:é|e)vrier)[.,;]?\s*\d+`),
	regexp.MustCompile(`(?i)\b[,;]?\s*(?:Mar|March|mars)[.,;]?\s*\d+`),
	regexp.MustCompile(`(?i)\b[,;]?\s*(?:Apr|April|avril)[.,;]?\s*\d+`),
	regexp.MustCompile(`(?i)\b[,;]?\s*(?:May|Mai)[.,;]?\s*\d+`),
	regexp.MustCompile(`(?i)\b[,;]?\s*(?:Jun|June|juin)[.,;]?\s*\d+`),
	regexp.MustCompile(`(?i)\b[,;]?\s*(?:Jul|July|juillet)[.,;]?\s*\d+`),
	regexp.MustCompile(`(?i)\b[,;]?\s*(?:Aug|August|ao(?:û|u)t)[.,;]?\s*\d+`),
	regexp.MustCompile(`(?i)\b[,;]?\s*(?:Sep|Sept|September|septembre)[.,;]?\s*\d+`),
	regexp.MustCompile(`(?i)\b[,;]?\s*(?:Oct|October|octobre)[.,;]?\s*\d+`),
	regexp.MustCompile(`(?i)\b[,;]?\s*(?:Nov|November|novembre)[.,;]?\s*\d+`),
	regexp.MustCompile(`(?i)\b[,;]?\s*(?:Dec|December|d(?:é|e)cembre)[.,;]?\s*\d+`),
	regexp.MustCompile(`(?i)\d+\s+(?:Jan|January|janvier)\.?\b`),
	regexp.MustCompile(`(?i)\d+\s+(?:Feb|February|f(?:é|e)vrier)\.?\b`),
	regexp.MustCompile(`(?i)\d+\s+(?:Mar|March|mars)\.?\b`),
	regexp.MustCompile(`(?i)\d+\s+(?:Apr|April|avril)\.?\b`),
	regexp.MustCompile(`(?i)\d+\s+(?:May|Mai)\b`),
	regexp.MustCompile(`(?i)\d+\s+(?:Jun|June|juin)\.?\b`),
	regexp.MustCompile(`(?i)\d+\s+(?:Jul|July|juillet)\.?\b`),
	regexp.MustCompile(`(?i)\d+\s+(?:Aug|August|ao(?:û|u)t)\.?\b`),
	regexp.MustCompile(`(?i)\d+\s+(?:Sep|September|septembre)t?\.?\b`),
	regexp.MustCompile(`(?i)\d+\s+(?:Oct|October|octobre)\.?\b`),
	regexp.MustCompile(`(?i)\d+\s+(?:Nov|November|novembre)\.?\b`),
	regexp.MustCompile(`(?i)\d+\s+(?:Dec|December|d(?:e|é)cembre)\.?\b`),
	regexp.MustCompile(`(?i)\b[-.,;:/]?\s*(?:Alabama|Alaska|Arizona|Arkansas|California|Colorado|Connecticut|Delaware|Evergreen|Florida|Hawaii|Idaho|Illinois|Indiana|Iowa|Kansas|Kentucky|Louisiana|Maine|Maryland|Massachusetts|Michigan|Minnesota|Mississippi|Missouri|Montana|Nebraska|Nevada|New\s+Hampshire|New\s+Jersey|New\s+Mexico|New\s+York|North\s+Carolina|North\s+Dakota|Ohio|Oklahoma|Oregon|Pennsylvania|Portland|Rhode\s+Island|South\s+Carolina|South\s+Dakota|St\s+Petersburg|Tennessee|Texas|Utah|Vermont|Washington|West\s+Virginia|Wisconsin|Wyoming)\s+State\s*\b`),
	regexp.MustCompile(`(?i)\b[.,;:/]?\s*(?:Afghanistan|Albania|Algeria|Australia|Austria|Azerbaijan|Bahamas|Bangladesh|Belarus|Belgium|Bolivia|Brazil|Bulgaria|Cambodia|Canada|Chile|China|Colombia|Croatia|Cuba|Cyprus|Denmark|Ecuador|Egypt|Finland|France|Germany|Ghana|Greece|Guatemala|Haiti|Hungary|Iceland|India|Indonesia|Iran|Iraq|Ireland|Israel|Italy|Jamaica|Japan|Kazakhstan|Kenya|Latvia|Lebanon|Libya|Lithuania|Luxembourg|Malaysia|Mexico|Mongolia|Morocco|Mozambique|Myanmar|Nepal|Netherlands|New\s+Zealand|Nicaragua|Nigeria|Norway|Oman|Pakistan|Panama|Paraguay|Peru|Philippines|Poland|Portugal|Qatar|Romania|Russia(?:n\s+Federation)?|Rwanda|Saudi\s+Arabia|Senegal|Serbia|Singapore|Slovakia|Slovenia|Somalia|South\s+Africa|South\s+Sudan|Spain|Sri\s+Lanka|Sudan|Sweden|Switzerland|Syria|Taiwan|Tanzania|Thailand|Tunisia|Turkey|Uganda|Ukraine|United\s+Arab\s+Emirates|United\s+Kingdom|United\s+States(?:\s+of\s+America)?|Uruguay|Uzbekistan|Venezuela|Vietnam|Yemen|Zambia|Zimbabwe)\b`),
	regexp.MustCompile(`(?i)autres?\s+de|probab|likely|possibl(?:e|y)|doubtful`),
	// "maybe" as a standalone suffix word — but not when it's part of a name
	// like "Maybee". Only strip when preceded by a space (not at start).
	regexp.MustCompile(`(?i)\s+maybe\b`),
	regexp.MustCompile(`(?i)\b\s*prob\.\s*\b`),
	// Strip leading "prob." at the very start of a segment too
	regexp.MustCompile(`(?i)^prob\.?\s+`),
	regexp.MustCompile(`(?i)\b\s*field\s*number`),
	regexp.MustCompile(`(?i)\b\s*(?:malaise|light|pitfall|pan|suction|lobster|actinic\s+light|cdc|fisherm(?:a|e)n)\s*trap\s*\b`),
	regexp.MustCompile(`(?i)\|\s*collector\s*(?:field\s*)?number.*$`),
	regexp.MustCompile(`(?i)\(?[,]?\s*(?:(?:local)?\s?collectors?|data\s*recorder|netter|(?:oper|prepar)ator)\(?s?\)?\.?:?`),
	regexp.MustCompile(`(?i)\b[-.,;:]?\s*(?:department|faculty)\s*(?:of)?\s*(?:entomology|biology|zoology)`),
	regexp.MustCompile(`(?i)fide:?\s*\b`),
	regexp.MustCompile(`(?i)first\s+name\s+unknown`),
	regexp.MustCompile(`(?i)game\s+dept\.?\s*\b`),
	regexp.MustCompile(`(?i)see\s+notes?\s*(?:inside)?`),
	regexp.MustCompile(`(?i)see\s+letter\s+enclosed`),
	regexp.MustCompile(`(?i)(?:by)?\s+correspondance`),
	regexp.MustCompile(`(?i)pers\.?\s*comm\.?`),
	regexp.MustCompile(`(?i)crossed\s+out`),
	regexp.MustCompile(`(?i)(?:ohne|keine)\s+angaben`),
	regexp.MustCompile(`(?i)\(?source\(?`),
	regexp.MustCompile(`(?i)according\s+to`),
	regexp.MustCompile(`(?i)lanuv\d+`),
	regexp.MustCompile(`(?i)\b\s*name\b`),
	regexp.MustCompile(`(?i)\b\s*lost\b`),
	regexp.MustCompile(`(?i)nswobs`),
	// Strip ORCID label and its digit string "0000-0001-2345-6789" together
	// so the dashes are removed before spaceDashSpaceRe fires.
	// The bare digit pattern \b\d{4}-\d{4}-\d{4}-\d{4}\b is handled earlier
	// (before \b\d{2,}\b) so it is not duplicated here.
	regexp.MustCompile(`(?i)ORCID[\s\d-]*`),
	regexp.MustCompile(`MRI[\s-]PAS`),
	regexp.MustCompile(`urn:qm\.qld\.gov\.au:collector`),
	regexp.MustCompile(`(?i)University\s+of\s+(?:Southern\s+)?California(?:,\s+Berkeley)?`),
	regexp.MustCompile(`(?i)field\s+museum\s+of\s+natural\s+history`),
	regexp.MustCompile(`(?i)american\s+museum\s+of\s+natural\s+history`),
	regexp.MustCompile(`(?i)The\s+Paleontological\s+Research\s+Institution`),
	regexp.MustCompile(`(?i)museums?\s+victoria`),
	regexp.MustCompile(`(?i)\b\s*(?:united\s+states|russia)\s*\b`),
	regexp.MustCompile(`(?i)revised|photograph|fruits\s+only`),
	regexp.MustCompile(`(?i)-?\s*sight\s+(?:identifi?cation|id)\.?\s*\b`),
	regexp.MustCompile(`(?i)-?\s*synonym(?:y|ie)`),
	regexp.MustCompile(`(?i)\b\s*\(?(?:fe)?male\)?\s*\b`),
	regexp.MustCompile(`(?i)\bto\s+(?:sub)?spp?\.?`),
	regexp.MustCompile(`(?i)nom\.?\s+rev\.?`),
	regexp.MustCompile(`\b(?:FNA|DAO|HUH|FDNMB|MNHN|PNI|USNM|ZMUC|CSIRO|ACAD|USGS|NAWQA)\b`),
	regexp.MustCompile(`(?i)(?:para|topo|syn|holo|allo|choro|eco|iso|isoepi|isopara|karyo|morpho|neo|mero|pala|paralecto|paraneo|photo|schizo)?types?:?`),
	regexp.MustCompile(`AFSC/POLISH\s+SORTING\s+CTR\.?`),
	regexp.MustCompile(`(?i)university|mus(?:e|é)um|exhibits?`),
	regexp.MustCompile(`(?i)uqam`),
	regexp.MustCompile(`(?i)sem\s+(?:colec?tor|data)`),
	regexp.MustCompile(`(?i)no\s+coll\.?(?:ector)?`),
	regexp.MustCompile(`(?i)not?\s+(?:name|date|details?|specific)?\s*(?:given|name|date|noted)`),
	regexp.MustCompile(`(?i)non?\s+specificato`),
	// Strip "year Month" or "Month year" trailing date fragments not caught above
	// e.g. "2006 May", "2006 may", "May 2013"
	regexp.MustCompile(`(?i)\b\d{4}\s+(?:Jan(?:uary)?|Feb(?:ruary)?|Mar(?:ch)?|Apr(?:il)?|May|Jun(?:e)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:t(?:ember)?)?|Oct(?:ober)?|Nov(?:ember)?|Dec(?:ember)?)\b`),
	regexp.MustCompile(`(?i)\b(?:Jan(?:uary)?|Feb(?:ruary)?|Mar(?:ch)?|Apr(?:il)?|May|Jun(?:e)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:t(?:ember)?)?|Oct(?:ober)?|Nov(?:ember)?|Dec(?:ember)?)\s+\d{4}\b`),
	// Strip "no disponible" (Spanish for not available)
	regexp.MustCompile(`(?i)\bno\s+disponible\b`),
	// Strip "& CI team", "& others", "& party" type trailing institutional suffixes
	regexp.MustCompile(`(?i)\s*&\s+[A-Z]{2,}\s+(?:team|group|party|crew|staff)\b`),
	// Strip "North Dakota State University" and similar "State University" combos
	regexp.MustCompile(`(?i)\b(?:North\s+Dakota|South\s+Dakota|Iowa|Ohio|Penn(?:sylvania)?|Michigan|Oregon|Arizona|Utah|Kansas|Colorado|Florida|Kentucky|Georgia|Virginia|Louisiana|Alabama|Mississippi|Tennessee|Indiana|Minnesota|Wisconsin|Oklahoma|Missouri|Arkansas|Nebraska|Wyoming|Montana|Idaho|Nevada|Vermont|Maine|Delaware|Hawaii|Alaska)\s+State\s+University\b`),
	// Strip trailing month name that is a date artifact left after digit stripping.
	// Uses a custom replacer (see trailingMonthReplacer below) that keeps the
	// preceding word and removes only the month, so "Jan Jones Jan." → "Jan Jones"
	// but "Vlk, Jan" → unchanged (Jan follows comma, no preceding 3-letter word).
	trailingMonthStripRe,
	// Strip "checked:" with no space before name (colon acts as separator but
	// the word "checked" must also be removed from segment start)
	regexp.MustCompile(`(?i)^checked?\s*:`),
	// Strip trailing verb words at end of a segment (mirrors verb_start but for endings)
	// e.g. "C.E. Garton 1980 checked" after the colon splits off the rest
	regexp.MustCompile(`(?i)\s+(?:checked?|annotated?|confirmed?|verified?|det|redet|verif)\s*$`),
	// Strip "Jan. N" date pattern: month-abbrev followed by day number
	regexp.MustCompile(`(?i)\b(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sept?|Oct|Nov|Dec)\.?\s+\d+[,.]?\s*\d*$`),
	// Strip "N Month" at string start (e.g. "on 15 January" leaves "January")
	regexp.MustCompile(`(?i)^\d+\s+(?:Jan(?:uary)?|Feb(?:ruary)?|Mar(?:ch)?|Apr(?:il)?|May|Jun(?:e)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:t(?:ember)?)?|Oct(?:ober)?|Nov(?:ember)?|Dec(?:ember)?)\b`),
	// Strip "North Dakota State" (residual from "North Dakota State University" after
	// university strip, where only "State" is left and "North Dakota" precedes it)
	regexp.MustCompile(`(?i)\bNorth\s+Dakota\s+State\b`),
	// Strip standalone "State" when it follows a place name stripped by earlier patterns
	regexp.MustCompile(`(?i)^State$`),
	regexp.MustCompile(`[!@?]\s*-?\s*`),
	regexp.MustCompile(`[,;]$`),
	regexp.MustCompile(`^\w{0,2}$`),
	regexp.MustCompile(`^[A-Z]{2,}$`),
	regexp.MustCompile(`(?i)annot\.?\s*?\b`),
	// stet[!,] only at end of string or before a digit (year) — NOT mid-word.
	// "Kronenstet" should not match. Allow optional spaces before "!" too.
	regexp.MustCompile(`(?i)\s+stet[\s!,]*\d*$`),
	// NOTE: standalone month-at-end strip removed — it incorrectly stripped
	// given names like "Jan" in "Vlk, Jan". Month artifacts are handled
	// by the trailingMonthStripRe function (requires preceding 3-letter word).
	regexp.MustCompile(`(?i)\s+prep\.?\s*$`),
	regexp.MustCompile(`[({].*?[)}]`),
	regexp.MustCompile(`\s+\[[\w\s?.-]{10,}\]`),
	regexp.MustCompile(`[({][A-Za-z]{1,3}$`),
	regexp.MustCompile(`(?i)\bleg[.:]?(?:\s|$)`),
	regexp.MustCompile(`[Dd](?:ed|on)[.:]`),
	regexp.MustCompile(`\s+[A-Z]*\d+$`),
	regexp.MustCompile(`\s+\d+[A-Za-z]+$`),
	regexp.MustCompile(`^[-,.\s;*\d]+\s?`),
	regexp.MustCompile(`\s*-{2,}\s*`),
	regexp.MustCompile(`(?i)^exc?p?[:.]?\s*`),
	regexp.MustCompile(`(?i)^(?:ex\.?|in)\s+(?:he?r?b)\.?\s+`),
	regexp.MustCompile(`(?i)(?:ex\.?|in)\s+(?:he?r?b)\.?\s+.*$`),
	regexp.MustCompile(`(?i):?\s*exch(?:\b|$)`),
	regexp.MustCompile(`\s+de\s*$`),
	// Strip trailing isolated dot (e.g. after bracket removal: "Holm, E .")
	regexp.MustCompile(`\s+\.\s*$`),
	// Strip "not any" and "has not" as complete phrases (given-blacklist phrases
	// that appear as display-order names if not caught early).
	regexp.MustCompile(`(?i)^not\s+any$|^has\s+not$`),
	// Strip leading "of " left after "University of X" has university removed
	regexp.MustCompile(`(?i)^\s*of\s+.*$`),  // strip "of Michigan" etc after institution removed
	// Strip word+digit compounds like "Smith2" (short digit → whole token removed)
	// or "Smith12345" (long digit suffix → just the digits stripped, letters kept).
	// Short (1-3 digit suffix): strip whole compound so "Smith2" → "" → 0 names.
	regexp.MustCompile(`\b[A-Za-z]{2,}\d{1,3}\b`),
	// Long digit suffix (4+): a capturing sub handled in stripOut keeps the letters.
	regexp.MustCompile(`\.{2,}$`),
}

// trailingMonthStripRe matches a 3+-letter word followed by a month name at end
// of string. Group 1 captures the preceding word so it can be preserved.
var trailingMonthStripRe = regexp.MustCompile(
	`(?i)(\b[A-Za-z\x{00C0}-\x{017E}]{3,})\s+` +
		`(?:Jan(?:uary)?|Feb(?:ruary)?|Mar(?:ch)?|Apr(?:il)?|May|Jun(?:e)?|Jul(?:y)?|` +
		`Aug(?:ust)?|Sep(?:t(?:ember)?)?|Oct(?:ober)?|Nov(?:ember)?|Dec(?:ember)?|` +
		`janvier|f[eé]vrier|mars|avril|juin|juillet|ao[uû]t|septembre|` +
		`octobre|novembre|d[eé]cembre)\.?\s*$`)

// longDigitSuffixRe strips long digit runs (4+) attached to words, keeping the letters.
// e.g. "Smith12345" → "Smith". Short suffixes (1-3) are handled by a separate pattern
// in stripOutPatterns that removes the whole token.
var longDigitSuffixRe = regexp.MustCompile(`([A-Za-z]{2,})\d{4,}`)

func stripOut(s string) string {
	// Strip long digit suffixes first (keep letters): "Smith12345" → "Smith"
	s = longDigitSuffixRe.ReplaceAllString(s, "$1")
	for _, re := range stripOutPatterns {
		if re == trailingMonthStripRe {
			// Custom replacement: keep the preceding word (group 1), strip only the month.
			s = trailingMonthStripRe.ReplaceAllStringFunc(s, func(m string) string {
				sub := trailingMonthStripRe.FindStringSubmatch(m)
				if len(sub) >= 2 {
					return sub[1] // return just the preceding word
				}
				return ""
			})
			continue
		}
		s = re.ReplaceAllString(s, " ")
	}
	return s
}

// ─────────────────────────────────────────────────────────────────────────────
// POST_STRIP_TIDY
// ─────────────────────────────────────────────────────────────────────────────
var postStripTidyRe = regexp.MustCompile("^\\s*[&,;.]\\s*|[\\[\\]]|^[`'\".,!?]+|[`'\",]+$")

func postStripTidy(s string) string {
	return postStripTidyRe.ReplaceAllString(s, "")
}

// ─────────────────────────────────────────────────────────────────────────────
// Pipe splitter (primary separator after char-subs)
// ─────────────────────────────────────────────────────────────────────────────
var splitByPipeRe = regexp.MustCompile(`\s*\|\s*`)

// spaceDashSpaceRe matches Ruby's SPLIT_BY \s+-\s+ separator.
// Spaces on both sides are required so hyphenated names ("García-López")
// and attached suffixes ("-jr") are not broken.
// Replacement is " | " (a pipe) so it is treated as a name boundary by
// processComplexSeps and parseNames, matching Ruby's behaviour exactly.
var spaceDashSpaceRe = regexp.MustCompile(`\s+-\s+`)

// conjunctionSepRe splits on conjunction words used as name separators.
// Multi-letter conjunctions (and, et, och, etc.) are case-insensitive via (?i:...).
// Single-letter conjunctions (e, y, i) are case-SENSITIVE — they must be lowercase
// so that uppercase initials like "A Y Jackson" or "Jack E Smith" are never split.
// Go's RE2 supports (?i:...) inline flag groups.
var conjunctionSepRe = regexp.MustCompile(
	`(?i:\s+\b(?:con|en|et|or|per|for|und|and|with|och)\b\s+)` +
		`|\s+[eyi]\s+`)

// splitByVerbRe mirrors the verb/role phrases in Ruby's SPLIT_BY.
// Each phrase introduces a new agent in a collector chain.
// Note: no trailing \b — many phrases end in "." which is not a word char.
// A trailing \s+ is required, so the phrase must be surrounded by whitespace.
var splitByVerbRe = regexp.MustCompile(
	`(?i)\s+(?:` +
	`annotated?(?:\s+by)?|` +
	`checked?(?:\s*[:\s]by)?|` +
	`comm\.?|` +
	`communicat\w*(?:\s+to)?|` +         // covers "communicatd to" typo
	`conf\.?(?:\s+by)?|confirmed?(?:\s+by)?|` +
	`confirmada?(?:\s+por)?|` +
	`det\.?(?:\s+by)?|` +
	`(?:donated\s+)?by|` +
	`dupl?\.?(?:\s+by)?|duplicate(?:\s+by)?|` +
	`ex\.?(?:\s+by)?|examined?(?:\s+by)?|` +
	`in?dentified?(?:\s+by)?|` +
	`in\s+coll\.?|` +
	`in\s+part(?:\s+by)?|` +
	`per|` +
	`prep\.?(?:\s+by)?|` +
	`purchased?(?:\s+by)?|` +
	`redet\.?(?:\s+by)?|reidentified?(?:\s+by)?|` +
	`stet[!,]?\s*|` +                    // "stet" as mid-string separator
	`then(?:\s+by)?|` +
	`ver\.?(?:\s+by)?|veri?f?\.?:?(?:\s+by)?|` +
	`v(?:e|é)rifi(?:e|é)e?d?(?:\s+(?:by|par))?|v[eé]rifi[eé]\s*|` +
	`via|from` +
	`)\s+`)

// splitByVerbAtStartRe strips verb phrases that appear at the very beginning
// of a segment (before the first name), e.g. "via Serena Lowartz",
// "by P. Zika", "annotated Yves Archambault", "prep. C.J. Guiguet".
// These leave a single name rather than creating a split.
var splitByVerbAtStartRe = regexp.MustCompile(
	`(?i)^(?:` +
	`stet[!,]?|` +
	`annotated?(?:\s+by)?|` +
	`checked?(?:\s+by)?|` +
	`comm\.?|` +
	`communicat\w*(?:\s+to)?|` +
	`conf\.?(?:\s+by)?|confirmed?(?:\s+by)?|` +
	`det\.?(?:\s+by)?|` +
	`(?:donated\s+)?by|` +
	`dupl?\.?(?:\s+by)?|` +
	`ex\.?(?:\s+by)?|examined?(?:\s+by)?|` +
	`in?dentified?(?:\s+by)?|` +
	`in\s+coll\.?|` +
	`per|` +
	`prep\.?(?:\s+by)?|` +
	`redet\.?(?:\s+by)?|` +
	`ver\.?(?:\s+by)?|veri?f?\.?:?(?:\s+by)?|` +
	`v(?:e|é)rifi(?:e|é)e?d?(?:\s+(?:by|par))?|v[eé]rifi[eé]\s*|` +
	`via` +
	`)\s+`)

// splitByVerbAtEndRe strips verb/role words that appear at the end of a segment.
// e.g. "C.E. Garton checked" (after the colon-split strips the rest)
var splitByVerbAtEndRe = regexp.MustCompile(
	`(?i)\s+(?:` +
	`annotated?|checked?|confirmed?|det|redet|verif(?:ied)?|verified?` +
	`)\s*$`)

// splitByPunctuationRe covers the single-character and Unicode separators
// from Ruby's SPLIT_BY [–|ǀ∣｜│&+\/;:] plus the "a." Catalan separator
// and [;,]{2,} (multiple semicolons or commas).
// en-dash (–), colon (:), and double-separators were previously missing.
// Note: single & and ; are already handled earlier in the pipeline.
// Slash (/) is also a separator (e.g. "O.Bennedict/G.J. Spencer").
var splitByPunctuationRe = regexp.MustCompile(
	`\s+a\.\s+|` +
	`[;,]{2,}|` +
	`[–:]|` +
	`\s*/\s*`)

// ─────────────────────────────────────────────────────────────────────────────
// Complex separator substitutions (COMPLEX_SEPARATORS in Ruby).
// ─────────────────────────────────────────────────────────────────────────────
type complexSep struct {
	re          *regexp.Regexp
	replacement string
}

var complexSeparators = []complexSep{
	// "Mrs./Mr. & Mrs./Mr. Family" shared-appellation
	{regexp.MustCompile(`^(Mrs?\.?)\s+&\s+(Mrs?\.?)\s+(.*)$`), "$1 $3 | $2 $3"},
	// "Family, Mrs. Given" → "Mrs. Given Family"
	{regexp.MustCompile(`^(\S{4,}),\s+(Mrs?\.|MRS?\.)\s+([A-Za-z.\s]+)$`), "$2 $3 $1"},

	// Two sort-order names concatenated: "Puttock, C.F. James, S.A."
	// Relaxed to \S{2,} to also match short names like "Ng, J."
	{regexp.MustCompile(`^(\S{2,},\s+(?:\S\.\s*)+)\s+(\S{2,},\s+(?:\S\.\s*)+)$`), "$1 | $2"},

	// "Family Q., Family P." — two sort-order entries where given is a single dotted initial
	// e.g. "Groom Q., Desmet P." → "Groom, Q. | Desmet, P."
	{regexp.MustCompile(`^([A-Z][a-z]+)\s+([A-Z]\.),\s+([A-Z][a-z]+)\s+([A-Z]\.)$`), "$1, $2 | $3, $4"},

	// dot-then-word: "J.R.Smith" → "J.R. Smith" (split compact names)
	{regexp.MustCompile(`(\S\.)([[:alpha:]]{2,})`), "$1 $2"},

	// "FamilyName Initials" display-order: "Picard J.H." → "Picard, J.H."
	{regexp.MustCompile(`^([[:alpha:]]{2,})(?:\s+)((?:\S\.\s?)+)$`), "$1, $2"},

	// van/von particle double-name
	{regexp.MustCompile(`^([[:alpha:]]*),?\s*(.*)\s+(van|von|v\.|von\s+der|van\s+der)(?:and|&|et|e|,|;)\s*([[:alpha:]]*),?\s*(.*)\s+(van|von|v\.|von\s+der|van\s+der)$`), "$3 $1, $2 | $6 $4, $5"},
	{regexp.MustCompile(`^([[:alpha:]]*),?\s*(.*)\s+(van|von|v\.|von\s+der|van\s+der)$`), "$3 $1, $2"},

	// "F.G. and/& H.I. Family1 Family2" (different families)
	{regexp.MustCompile(`^((?:[A-Z]\.\s?)+)\s*(?:and|&|et|e)\s+((?:[A-Z]\.\s?)+)\s+([[:alpha:]'-]{2,})\s+([[:alpha:]'-]{2,})$`), "$1 $4 | $2 $3 $4"},
	// "F.G. and/& H.I. SharedFamily"
	{regexp.MustCompile(`^((?:[A-Z]\.\s?)+)\s*(?:and|&|et|e)\s+((?:[A-Z]\.\s?)+)\s+([[:alpha:]'-]{2,})(.*)$`), "$1 $3 | $2 $3 | $4"},
	// "A and/& B SharedFamily" (single-letter initials without dots)
	{regexp.MustCompile(`^([A-Z]{1,3})\s*(?:and|&|et|e)\s+([A-Z]{1,3})\s+([[:alpha:]'-]{2,})(.*)$`), "$1 $3 | $2 $3 | $4"},

	// "F.G., H.I. and/& J.K. SharedFamily" (3-way shared family)
	{regexp.MustCompile(`^((?:[A-Z]\.\s?)+),\s+([A-Z.\s]+)\s+(?:and|&|et|e)\s+((?:[A-Z]\.\s?)+)\s+([[:alpha:]'-]{2,})(.*)$`), "$1 $4 | $2 $4 | $3 $4 | $5"},

	// Word-list separators (Chaboo, Bennett, Shin style — all family-only names)
	{regexp.MustCompile(`(?i)^([A-Z][[:alpha:]]{2,}),\s*([A-Z][[:alpha:]]{2,})\s*(?:and|&|et|e|,)\s+([A-Z][[:alpha:]]{2,})$`), "$1 | $2 | $3"},
	{regexp.MustCompile(`(?i)^([A-Z][[:alpha:]]{2,}),\s*([A-Z][[:alpha:]]{2,}),\s*([A-Z][[:alpha:]]{2,})\s*(?:and|&|et|e|,)\s+([A-Z][[:alpha:]]{3,})$`), "$1 | $2 | $3 | $4"},
	{regexp.MustCompile(`(?i)^([A-Z][[:alpha:]]{2,}),\s*([A-Z][[:alpha:]]{2,}),\s*([A-Z][[:alpha:]]{2,}),\s*([A-Z][[:alpha:]]{2,})\s*(?:and|&|et|e|,)\s+([A-Z][[:alpha:]]{3,})$`), "$1 | $2 | $3 | $4 | $5"},
}

// displayOrderListItemRe matches a single display-order name item.
// Handles:
//   "N. Lujan"       — single initial + family
//   "C.F. James"     — multi-initial + family
//   "C Armbruster"   — undotted initial + family
//   "Guy C. Joslin"  — full given word(s) + initial + family
//   "Jean-Pierre Blanc" — hyphenated given + family
// Does NOT match pure-initials like "S.A." or bare single letters like "K."
var displayOrderListItemRe = regexp.MustCompile(
	`^` +
		`(?:[A-Za-z][A-Za-z\x{00C0}-\x{017E}\x{02B0}-\x{036F}'-]+\s+)*` + // optional full given words (incl. hyphenated)
		`(?:` +
		`(?:[A-Z]\.)+\s+` + // dotted multi-initial prefix: "C.F. " or "W.J.K. "
		`|[A-Z]\.?\s+` + // single letter with optional dot: "N. " or "C "
		`)*` +
		`[A-Z][a-z\x{00E0}-\x{017E}][^\s,]{0,}$`) // family: uppercase + lowercase start

// displayOrderListInitialItemRe matches items that have an initials prefix —
// a single initial ("N. Lujan"), multiple dotted initials ("C.H. Lowe"),
// or a full given word followed by initials ("Guy C. Joslin").
// This is used as the has_i guard: at least one item in the comma list must
// match this to distinguish display-order lists from bare family-name lists.
var displayOrderListInitialItemRe = regexp.MustCompile(
	`^(?:[A-Za-z][A-Za-z\x{00C0}-\x{017E}'-]+\s+)*(?:[A-Z]\.)+\s+[A-Z]` + // "Guy C. J..." or "C.H. L..."
		`|^[A-Z]\.?\s+[A-Z]`) // "P. W..." or "N. L..."

func isDisplayOrderList(seg string) bool {
	if !strings.Contains(seg, ",") {
		return false
	}
	parts := strings.Split(seg, ",")
	// Clean and filter
	cleaned := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.TrimLeft(p, "& \t")
		if p != "" {
			cleaned = append(cleaned, p)
		}
	}
	if len(cleaned) < 2 {
		return false
	}
	matched := 0
	hasInitialsPrefix := false
	for _, p := range cleaned {
		if displayOrderListItemRe.MatchString(p) {
			matched++
		}
		if displayOrderListInitialItemRe.MatchString(p) {
			hasInitialsPrefix = true
		}
	}
	// For small lists (≤3 items): require ALL items match AND at least one has an
	// initials prefix. This prevents "Puttock, C.F. James, S.A." (two concatenated
	// sort-order names) from being misidentified as a display-order list.
	// For larger lists (≥4 items): require at least 2/3 matching with an initials prefix.
	if len(cleaned) <= 3 {
		return matched == len(cleaned) && hasInitialsPrefix
	}
	return matched >= 2 && matched*3 >= len(cleaned)*2 && hasInitialsPrefix
}

// sortOrderListItemRe matches a sort-order name token: "Harkness, W.J.K."
// In the comma-list context, each token is "Family" and the following
// token is the initials — so after splitting on ", " we get pairs.
var sortOrderListItemRe = regexp.MustCompile(
	`^[A-Z][a-z]{1,}$`)   // just the family part token
var initialsOnlyRe = regexp.MustCompile(
	`^(?:[A-Z]\.)+$`)     // just the initials part token

// isSortOrderList detects "Harkness, W.J.K., Dickinson, J.C., & Marshall, N."
// A sort-order list has the pattern: Word, Initials, Word, Initials...
func isSortOrderList(seg string) bool {
	// Quick check: contains multiple commas and starts with a capital word
	if !strings.Contains(seg, ",") {
		return false
	}
	parts := strings.Split(seg, ",")
	if len(parts) < 4 { // need at least "Family, Init, Family, Init"
		return false
	}
	// Check pattern: odd parts are family names (single word), even are initials
	matchCount := 0
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if i%2 == 0 {
			if sortOrderListItemRe.MatchString(p) {
				matchCount++
			}
		} else {
			if initialsOnlyRe.MatchString(p) || regexp.MustCompile(`^[A-Z]\.`).MatchString(p) {
				matchCount++
			}
		}
	}
	return matchCount >= 4
}

func splitSortOrderList(seg string) string {
	// Split "Harkness, W.J.K., Dickinson, J.C., & Marshall, N." into pipe-list.
	// The charSubs rule ", &" → " &" can remove the comma before "&", merging
	// the initials and the next family name: "J.C., & Marshall" → "J.C. & Marshall".
	// We handle that by splitting on " & " within the initials token.
	initRe := regexp.MustCompile(`^[A-Z]\.`)
	parts := strings.Split(seg, ",")
	var names []string
	i := 0
	for i < len(parts) {
		fam := strings.TrimSpace(strings.TrimLeft(parts[i], "& \t"))
		if fam == "" {
			i++
			continue
		}
		if i+1 < len(parts) {
			giv := strings.TrimSpace(strings.TrimLeft(parts[i+1], "& \t"))
			// Handle "J.C. & Marshall" — initials merged with next family via charSubs
			if ampIdx := strings.Index(giv, " & "); ampIdx >= 0 {
				actualGiv := strings.TrimSpace(giv[:ampIdx])
				nextFam := strings.TrimSpace(giv[ampIdx+3:])
				if initRe.MatchString(actualGiv) && sortOrderListItemRe.MatchString(nextFam) {
					names = append(names, fam+", "+actualGiv)
					// Peek at next part for nextFam's given
					if i+2 < len(parts) {
						nextGiv := strings.TrimSpace(strings.TrimLeft(parts[i+2], "& \t"))
						if initRe.MatchString(nextGiv) {
							names = append(names, nextFam+", "+nextGiv)
							i += 3
							continue
						}
					}
					names = append(names, nextFam)
					i += 2
					continue
				}
			}
			if giv != "" && (initialsOnlyRe.MatchString(giv) || initRe.MatchString(giv)) {
				names = append(names, fam+", "+giv)
				i += 2
				continue
			}
		}
		if sortOrderListItemRe.MatchString(fam) {
			names = append(names, fam)
		}
		i++
	}
	if len(names) >= 2 {
		return strings.Join(names, " | ")
	}
	return seg
}

func applyComplexSeparators(s string) string {
	// Sort-order list: "Harkness, W.J.K., Dickinson, J.C., & Marshall, N."
	if isSortOrderList(s) {
		return splitSortOrderList(s)
	}

	// Display-order comma list detection. charSubs converts ", &" → " &"
	// (losing the comma), so also try with " & " normalised back to ",".
	sNorm := regexp.MustCompile(`\s*&\s*`).ReplaceAllString(s, ",")
	if isDisplayOrderList(sNorm) {
		parts := strings.Split(sNorm, ",")
		trimmed := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				trimmed = append(trimmed, p)
			}
		}
		return strings.Join(trimmed, " | ")
	}

	for _, cs := range complexSeparators {
		if cs.re.MatchString(s) {
			s = cs.re.ReplaceAllString(s, cs.replacement)
		}
	}
	return s
}
// ─────────────────────────────────────────────────────────────────────────────
// BLACKLIST regex (applied to DisplayOrder of a parsed name)
// ─────────────────────────────────────────────────────────────────────────────
var blacklistRe = regexp.MustCompile(`(?i)` +
	`\bherb\b|` +
	`abundant|` +
	`adult|juvenile|` +
	`administra(?:d|t)or|` +
	`australian?|` +
	`average|` +
	`believe|unclear|ill?egible|suggested|disagrees?|` +
	`\bnone\b|` +
	`barcod|bgwd|` +
	`(?:biolog|botan|zoo|ecolog|mycol|(?:in)?vertebrate|fisheries|genetic|animal|mushroom|wildlife|plumage|flower|agriculture)|` +
	`(?:bri?tish|canadi?an?|chinese|arctic|japan|russian|north\s+america)|` +
	`carex|salix|` +
	`catalog(?:ue)?|` +
	`conservator|` +
	`(?:herbarium|herbier|collection|collected|publication|specimen|species|describe|an(?:a|o)morph|isolated|recorded|inspection|define|status|lighthouse)|` +
	`\bhelp\b|` +
	`data\s+not\s+captured|` +
	`(?:description|drawing|identification|remark|original|illustration|checklist|intermedia|measurement|indisting|series|imperfect)|` +
	`desconocido|` +
	`exc(?:s?icc?at(?:a|i))|` +
	`evidence|exporter|foundation|` +
	`ichthyology|inconn?u|` +
	`(?:internation|gou?vern|ministry|extension|unit|district|provincial|na(?:c|t)ional|military|region|environ|natur(?:e|al)|naturelles|division|program|direction)|` +
	`label|o\.?m\.?n\.?r\.?|measurement|ent(?:o|y)mology|malacology|geographic|` +
	`(?:mus(?:eum|ée)|universit(?:y|é|e|at)|college|institute?|acad(?:e|é)m|school|laboratoi?r|project|polytech|dep(?:t|artment)|research|clinic|hospital|cientifica|sanctuary|safari)|` +
	`univ\.|` +
	`\b(?:graduate|student|storekeep|supervisor|superint|rcmp|coordinator|minority|fisherm(?:a|e)n|police|taxonomist|consultant|team|crew|group|personnel|staff|family|captain|friends|assistant|worker|gamekeeper)\b|` +
	`non\s+pr(?:é|e)cis(?:é|e)|no\s+consta|` +
	`no\s+(?:agent\s+)?(?:data|disponible)(?:\s+available)?|` +
	`not?\s+(?:entered|stated)|nomenclatur(?:e|al)\s+adjustment|not\s+available|` +
	`(?:ontario|qu(?:e|é)bec|saskatchewan|new\s+brunswick|sault|newfoundland|assurance|vancouver|u\.?s\.?s\.?r\.?)|` +
	`popa\s+observers?|recreation|culture|renseigné|` +
	`(?:shaped|dark|pale|areas|phase|spotting|interior|between|closer)|` +
	`soci(?:e|é)t(?:y|é)|cent(?:er|re)|community|history|conservation|conference|assoc|commission|consortium|council|club|exposit|alliance|protective|circle|` +
	`commercial|control|product|sequence\s+data|size|large|colou?r|skeleton|` +
	`survey|assessment|station|monitor|stn\.|engine|(?:e|é)x?chang(?:e|é)s?|ex(?:c|k)urs(?:e|o|ó)n?|exped\.?|exp(?:e|i)di(?:c|t)i(?:e|o|ó)n?|experiment|explora(?:d|t)|festival|generation|inventory|marine|service|` +
	`submersible|synonymy?|systematic|perspective|` +
	`taxiderm(?:ies|y)|though|texas\s+instruments?(?:\s+for)?|tropical|toward|seen\s+at|` +
	`unidentified|unspecified|unk?nown?|unnamed|unread|unmistak|no\s+agent|` +
	`urn:|usda|ucla|workshop|garden|farm|jardin|public`,
)

// ─────────────────────────────────────────────────────────────────────────────
// Lists and maps
// ─────────────────────────────────────────────────────────────────────────────

var familyGreenlist = map[string]bool{
	"ng": true, "srb": true, "srp": true, "vlk": true,
	"smrz": true, "smrž": true, "smrt": true, "krc": true, "krč": true,
}

var familyBlacklist = map[string]bool{
	"a b": true, "a e": true, "a g": true, "a j": true, "a k": true,
	"ap": true, "da": true, "de": true, "de'": true, "del": true,
	"der": true, "di": true, "do": true, "dos": true, "du": true,
	"el": true, "la": true, "nebc": true, "van": true, "von": true,
	"the": true, "of": true, "new": true, "no": true, "pp": true,
	"adjustment": true, "agent": true, "annotated": true, "annotator": true,
	"available": true, "arachnology": true, "catalogue": true,
	"checked": true, "collector": true, "comments": true, "confirmed": true,
	"curators": true, "data": true, "details": true, "determiner": true,
	"determination": true, "dissected": true, "dissection": true,
	"entered": true, "erased": true, "expd": true, "expdn": true,
	"hist": true, "historical": true, "historie": true,
	"indecipherable": true, "inst": true, "meteorological": true,
	"nomenclatural": true, "orig": true, "prep": true, "prof": true,
	"professional": true, "qld": true, "registration": true,
	"science": true, "state": true, "stet": true, "study": true, "umeå": true,
	"verified": true, "wg": true, "wm": true, "wn": true, "zw": true,
	"zz": true, "z-": true,
}

var givenBlacklist = map[string]bool{
	"not any": true, "has not": true,
}

var particles = map[string]bool{
	"ap": true, "da": true, "de": true, "de'": true, "del": true,
	"der": true, "des": true, "di": true, "do": true, "dos": true,
	"du": true, "el": true, "le": true, "la": true, "van": true,
	"von": true, "the": true, "of": true,
	"van de": true, "van der": true, "von der": true,
}

// ─────────────────────────────────────────────────────────────────────────────
// Helper regexes and vars used across files
// ─────────────────────────────────────────────────────────────────────────────
var (
	dotThenWordRe         = regexp.MustCompile(`([A-Za-z]\.)([A-Za-z]{2,})`)
	trailingInitialRe     = regexp.MustCompile(`[A-Z]$`)
	residualTerminatorsRe = regexp.MustCompile(`[,;]\s*$`)
	vowels                = "aeiouàáâäǎæãåāèéêëěẽēėęìíîïǐĩīıįòóôöǒœøõōùúûüǔũūűů"
)

func containsVowel(s string) bool {
	for _, r := range strings.ToLower(s) {
		if strings.ContainsRune(vowels, r) {
			return true
		}
	}
	return false
}

func isUpperOrLower(s string) bool {
	if s == "" {
		return false
	}
	return s == strings.ToUpper(s) || s == strings.ToLower(s)
}