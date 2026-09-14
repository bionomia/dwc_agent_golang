package dwcagent_test

// Benchmark suite comparing Parse+Clean throughput.
// Run with:
//   go test -bench=. -benchtime=10s -benchmem
//
// To get ns/op for a realistic mix of inputs (the most meaningful number
// for estimating wall-clock time on 1M rows), use BenchmarkParseCleanMix.

import (
	"testing"

	dwcagent "github.com/bionomia/dwc_agent_golang"
)

// Representative sample of real-world recordedBy strings from GBIF,
// covering the main categories of complexity found in the wild.
var benchInputs = []string{
	// Simple display-order
	"W.J. Cody",
	"R.D.M. Page",
	"J. Smith",
	"Maria García",
	// Sort-order
	"Cody, W.J.",
	"Smith, John",
	"Müller, H.",
	// Multiple names, semicolon-separated
	"Smith, J.; Jones, A.; Brown, K.",
	"W.J. Cody; R.D.M. Page; H. Müller",
	// Numbers and noise (common in GBIF verbatim)
	"13267 (male) W.J. Cody; 13268 (female) W.E. Kemp",
	"leg. A. Rocabruna - M. Tabarés- J. Vila., Herb. SCM2498, (RIPOLLèS .)",
	// Conjunction separators
	"A. Gòmez-Bolea i A. Longàn",
	"Smith et Jones",
	"Wagner und Mueller",
	// Shared family name patterns
	"J. & K. Smith",
	"J. et K. Smith",
	// Role/verb separators
	"Smith, J. det. Jones, A.",
	"Smith, J. identified by Jones, A.",
	"Smith confirmed by Jones",
	// ORCID
	"Smith, J. ORCID 0000-0001-2345-6789",
	// Particles
	"Ludwig von Beethoven",
	"Jan van der Berg",
	// Titles and appellations
	"Sir Isaac Newton",
	"Dr. Smith, J.",
	"Ms. Sofia Kovaleskaya",
	// All noise — should return nothing
	"Anonymous",
	"Unknown",
	"N/A",
	"University of Michigan",
	"",
	// Long multi-collector strings
	"W.J. Cody; R.D.M. Page; H. Müller; A. Smith; B. Jones; C. Brown",
	"Of & Dr. L. & Dr. A. & Copenhagen, I",
	// Mojibake / corrupted data
	"8:>;0O 8E08;>28G0",
	// Sort-order with et al
	"Cody, W.J. et al.",
	// Hyphenated names
	"García-López, J.",
	"O'Brien, Patrick",
}

// BenchmarkParseOnly measures just the Parse step (no Clean).
func BenchmarkParseOnly(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := benchInputs[i%len(benchInputs)]
		_ = dwcagent.Parse(input)
	}
}

// BenchmarkParseAndClean measures Parse + Clean on every parsed name —
// the full pipeline that produces Agent records.
func BenchmarkParseAndClean(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := benchInputs[i%len(benchInputs)]
		names := dwcagent.Parse(input)
		for _, n := range names {
			_ = dwcagent.Clean(n)
		}
	}
}

// BenchmarkParseCleanMix is the headline number: how long does it take to
// process one "average" input string end-to-end?
// Divide 1,000,000,000 by the ns/op figure to get throughput in strings/sec,
// then divide 1M by that to get estimated wall-clock seconds for 1M strings.
func BenchmarkParseCleanMix(b *testing.B) {
	n := len(benchInputs)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := benchInputs[i%n]
		names := dwcagent.Parse(input)
		for _, name := range names {
			c := dwcagent.Clean(name)
			_ = c.IsDefault()
		}
	}
}

// BenchmarkParseCleanSimple measures a simple single-name string —
// the best case, closest to the Ruby gem's fast path.
func BenchmarkParseCleanSimple(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		names := dwcagent.Parse("W.J. Cody")
		for _, n := range names {
			_ = dwcagent.Clean(n)
		}
	}
}

// BenchmarkParseCleanComplex measures a noisy multi-name string —
// the expensive case.
func BenchmarkParseCleanComplex(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		names := dwcagent.Parse("13267 (male) W.J. Cody; 13268 (female) W.E. Kemp")
		for _, n := range names {
			_ = dwcagent.Clean(n)
		}
	}
}

// BenchmarkParseCleanNoisy measures a string that produces no valid names —
// common in GBIF data (institutions, anonymous, corrupted).
func BenchmarkParseCleanNoisy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		names := dwcagent.Parse("leg. A. Rocabruna - M. Tabarés- J. Vila., Herb. SCM2498, (RIPOLLèS .)")
		for _, n := range names {
			_ = dwcagent.Clean(n)
		}
	}
}

// Benchmark1MStrings gives a direct wall-clock estimate for 1M strings
// by running exactly 1,000,000 iterations of the mix benchmark.
// Run with: go test -bench=Benchmark1MStrings -benchtime=1x
func Benchmark1MStrings(b *testing.B) {
	n := len(benchInputs)
	b.ReportAllocs()
	for i := 0; i < 1_000_000; i++ {
		input := benchInputs[i%n]
		names := dwcagent.Parse(input)
		for _, name := range names {
			_ = dwcagent.Clean(name)
		}
	}
	b.N = 1_000_000
}
