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
	regexp.MustCompile(`\d*[A-Za-z]*\d*-\d*$`),
	regexp.MustCompile(`\b\d+\(?[[:alpha:]]\)?\b`),
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
	regexp.MustCompile(`(?i)\b[,;]?\s*string\b`),
	regexp.MustCompile(`(?i)\b[,;]?\s*person\s*string\b`),
	regexp.MustCompile(`(?i)^colln?\.?\s+|\s*colln?\.?\s*$`),
	regexp.MustCompile(`(?i)^collection:?\s+|\s*collection\s*$`),
	regexp.MustCompile(`(?i)\b[,;]?\s*colls\.(?:\b|$)`),
	regexp.MustCompile(`(?i)contactid`),
	regexp.MustCompile(`(?i)^dupl[.,]+`),
	regexp.MustCompile(`(?i)\b[,;]?\s*stet[,!]?\s*\d*$`),
	regexp.MustCompile(`(?i)[,;]?\s*\d+[-/\s](?:\d+|Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sept?|Oct|Nov|Dec)\.?\s*[-/\s]?\d+`),
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
	regexp.MustCompile(`(?i)\b\s*maybe\s*\b`),
	regexp.MustCompile(`(?i)\b\s*prob\.\s*\b`),
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
	regexp.MustCompile(`(?i)-?\s*sight\s+(?:id|identifi?cation)\.?\s*\b`),
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
	regexp.MustCompile(`\b[,;]\s+\d+\.?$`),
	regexp.MustCompile(`[!@?]\s*-?\s*`),
	regexp.MustCompile(`(?i)\d{1,4}[/.]?(?:i|ii|iii|iv|v|vi|vii|viii|ix|x|xi|xii)[/.]\d{1,4}`),
	regexp.MustCompile(`[,;]$`),
	regexp.MustCompile(`^\w{0,2}$`),
	regexp.MustCompile(`^[A-Z]{2,}$`),
	regexp.MustCompile(`(?i)annot\.?\s*?\b`),
	regexp.MustCompile(`(?i)\s+stet\s*!?\s*$`),
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
	regexp.MustCompile(`\.{2,}$`),
}

func stripOut(s string) string {
	for _, re := range stripOutPatterns {
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

// conjunctionSepRe splits on conjunction words used as name separators,
// mirroring Ruby's SPLIT_BY \b(con|e|y|i|en|et|or|per|for|und)\b.
// Spaces on both sides are required so particles like "den", "van", "von"
// and name fragments like "en" inside "Anderson" are not split.
// Must be applied AFTER processComplexSeps so shared-family patterns like
// "J. et K. Smith" are expanded before "et" becomes a pipe separator.
var conjunctionSepRe = regexp.MustCompile(`(?i)\s+\b(con|e|y|i|en|et|or|per|for|und|and|with|och)\b\s+`)

// splitByVerbRe mirrors the verb/role phrases in Ruby's SPLIT_BY.
// Each phrase introduces a new agent in a collector chain.
// Note: no trailing \b — many phrases end in "." which is not a word char.
// A trailing \s+ is required, so the phrase must be surrounded by whitespace.
var splitByVerbRe = regexp.MustCompile(
	`(?i)\s+(?:` +
	`annotated(?:\s+by)?|` +
	`checked?(?:\s+by)?|` +
	`comm\.?|` +
	`communicate?d(?:\s+to)?|` +
	`conf\.?(?:\s+by)?|confirmed(?:\s+by)?|` +
	`confirmada(?:\s+por)?|` +
	`det\.?(?:\s+by)?|` +
	`(?:donated\s+)?by|` +
	`dupl?\.?(?:\s+by)?|duplicate(?:\s+by)?|` +
	`ex\.?(?:\s+by)?|examined(?:\s+by)?|` +
	`in?dentified(?:\s+by)?|` +
	`in\s+coll\.?|` +
	`in\s+part(?:\s+by)?|` +
	`prep\.?(?:\s+by)?|` +
	`purchased?(?:\s+by)?|` +
	`redet\.?(?:\s+by)?|` +
	`reidentified(?:\s+by)?|` +
	`then(?:\s+by)?|` +
	`veri?f?\.?:?(?:\s+by)?|` +
	`v(?:e|é)rifi(?:e|é)e?d?(?:\s+(?:by|par))?|` +
	`via|from` +
	`)\s+`)

// splitByPunctuationRe covers the single-character and Unicode separators
// from Ruby's SPLIT_BY [–|ǀ∣｜│&+\/;:] plus the "a." Catalan separator
// and [;,]{2,} (multiple semicolons or commas).
// en-dash (–), colon (:), and double-separators were previously missing.
// Note: single & and ; are already handled earlier in the pipeline.
var splitByPunctuationRe = regexp.MustCompile(
	`\s+a\.\s+|` +
	`[;,]{2,}|` +
	`[–:]`)

// ─────────────────────────────────────────────────────────────────────────────
// Complex separator substitutions (COMPLEX_SEPARATORS in Ruby).
// ─────────────────────────────────────────────────────────────────────────────
type complexSep struct {
	re          *regexp.Regexp
	replacement string
}

var complexSeparators = []complexSep{
	{regexp.MustCompile(`^(\S{4,}),\s+(Mrs?\.|MRS?\.)\s+([A-Za-z.\s]+)$`), "$2 $3 $1"},
	{regexp.MustCompile(`^(Mrs?\.?)\s+&\s+(Mrs?\.?)\s+(.*)$`), "$1 $3 | $2 $3"},
	{regexp.MustCompile(`^([A-Z]\.[[:alpha:]]+),\s*([A-Z.]+)$`), "$1 $2"},
	{regexp.MustCompile(`^(\S{4,},\s+(?:\S\.\s*)+)\s+(\S{4,},\s+(?:\S\.\s*)+)$`), "$1 | $2"},
	{regexp.MustCompile(`(\S\.)([[:alpha:]]{2,})`), "$1 $2"},
	{regexp.MustCompile(`^([[:alpha:]]{2,})(?:\s+)((?:\S\.\s?)+)$`), "$1, $2"},
	{regexp.MustCompile(`^([[:alpha:]]*),?\s*(.*)\s+(van|von|v\.|von\s+der|van\s+der)(?:and|&|et|e|,|;)\s*([[:alpha:]]*),?\s*(.*)\s+(van|von|v\.|von\s+der|van\s+der)$`), "$3 $1, $2 | $6 $4, $5"},
	{regexp.MustCompile(`^([[:alpha:]]*),?\s*(.*)\s+(van|von|v\.|von\s+der|van\s+der)$`), "$3 $1, $2"},
	{regexp.MustCompile(`^((?:[A-Z]\.\s?)+)\s?(?:and|&|et|e)\s+((?:[A-Z]\.\s?)+)\s+([[:alpha:]'-]{2,})\s+([[:alpha:]'-]{2,})$`), "$1 $4 | $2 $3 $4"},
	{regexp.MustCompile(`^((?:[A-Z]\.\s?)+)\s?(?:and|&|et|e)\s+((?:[A-Z]\.\s?)+)\s+([[:alpha:]'-]{2,})(.*)$`), "$1 $3 | $2 $3 | $4"},
	{regexp.MustCompile(`^([A-Z]{1,3})\s+(?:and|&|et|e)\s+([A-Z]{1,3})\s+([[:alpha:]'-]{2,})(.*)$`), "$1 $3 | $2 $3 | $4"},
	{regexp.MustCompile(`^((?:[A-Z]\.\s?)+),\s+([A-Z.\s]+)\s+(?:and|&|et|e)\s+((?:[A-Z]\.\s?)+)\s+([[:alpha:]'-]{2,})(.*)$`), "$1 $4 | $2 $4 | $3 $4 | $5"},
	{regexp.MustCompile(`(?i)^([A-Z][[:alpha:]]{2,}),\s*([A-Z][[:alpha:]]{2,})\s*(?:and|&|et|e|,)\s+([A-Z][[:alpha:]]{2,})$`), "$1 | $2 | $3"},
	{regexp.MustCompile(`(?i)^([A-Z][[:alpha:]]{2,}),\s*([A-Z][[:alpha:]]{2,}),\s*([A-Z][[:alpha:]]{2,})\s*(?:and|&|et|e|,)\s+([A-Z][[:alpha:]]{3,})$`), "$1 | $2 | $3 | $4"},
	{regexp.MustCompile(`(?i)^([A-Z][[:alpha:]]{2,}),\s*([A-Z][[:alpha:]]{2,}),\s*([A-Z][[:alpha:]]{2,}),\s*([A-Z][[:alpha:]]{2,})\s*(?:and|&|et|e|,)\s+([A-Z][[:alpha:]]{3,})$`), "$1 | $2 | $3 | $4 | $5"},
}

func applyComplexSeparators(s string) string {
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
	"adjustment": true, "agent": true, "annotator": true, "available": true,
	"arachnology": true, "catalogue": true, "comments": true, "curators": true,
	"data": true, "details": true, "determiner": true, "determination": true,
	"dissected": true, "dissection": true, "entered": true, "erased": true,
	"expd": true, "expdn": true, "hist": true, "historical": true,
	"historie": true, "indecipherable": true, "inst": true,
	"meteorological": true, "nomenclatural": true, "orig": true,
	"prof": true, "professional": true, "qld": true, "registration": true,
	"science": true, "study": true, "umeå": true, "wg": true, "wm": true,
	"wn": true, "zw": true, "zz": true, "z-": true,
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
