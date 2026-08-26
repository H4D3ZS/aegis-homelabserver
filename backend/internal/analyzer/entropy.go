package analyzer

import (
	"math"
	"strings"
)

// ShannonEntropy calculates the information entropy of a string: H(X) = -sum(P(x) * log2(P(x)))
func ShannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0.0
	}

	freq := make(map[rune]float64)
	for _, r := range s {
		freq[r]++
	}

	length := float64(len(s))
	var entropy float64

	for _, count := range freq {
		p := count / length
		entropy -= p * math.Log2(p)
	}

	return entropy
}

// NGramEntropy calculates 2-gram transition entropy of a string.
func NGramEntropy(s string) float64 {
	if len(s) < 2 {
		return 0.0
	}

	bigrams := make(map[string]float64)
	totalBigrams := float64(len(s) - 1)

	for i := 0; i < len(s)-1; i++ {
		bigram := s[i : i+2]
		bigrams[bigram]++
	}

	var entropy float64
	for _, count := range bigrams {
		p := count / totalBigrams
		entropy -= p * math.Log2(p)
	}

	return entropy
}

// ExtractSLD returns the second-level domain from an FQDN.
func ExtractSLD(domain string) string {
	domain = strings.TrimSuffix(domain, ".")
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return domain
	}
	return parts[len(parts)-2]
}
