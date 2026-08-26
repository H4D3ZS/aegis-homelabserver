package analyzer

import (
	"strings"
	"unicode"
)

// ThreatResult represents the detailed heuristic evaluation of a DNS query.
type ThreatResult struct {
	Domain          string   `json:"domain"`
	SLD             string   `json:"sld"`
	ShannonEntropy  float64  `json:"shannon_entropy"`
	NGramEntropy    float64  `json:"ngram_entropy"`
	ConsecutiveCons int      `json:"consecutive_consonants"`
	VowelRatio      float64  `json:"vowel_ratio"`
	HexDensity      float64  `json:"hex_density"`
	ThreatScore     float64  `json:"threat_score"` // 0.0 - 1.0
	IsThreat        bool     `json:"is_threat"`    // true if ThreatScore >= Threshold or Shannon >= 3.85
	Indicators      []string `json:"indicators"`
}

// DGAAnalyzer evaluates domains for DGA (Domain Generation Algorithms) & DNS tunneling patterns.
type DGAAnalyzer struct {
	ThreatThreshold float64
}

// NewDGAAnalyzer creates a new analyzer instance with default 0.75 threat threshold.
func NewDGAAnalyzer(threshold float64) *DGAAnalyzer {
	if threshold <= 0.0 {
		threshold = 0.75
	}
	return &DGAAnalyzer{
		ThreatThreshold: threshold,
	}
}

func isVowel(r rune) bool {
	switch unicode.ToLower(r) {
	case 'a', 'e', 'i', 'o', 'u', 'y':
		return true
	default:
		return false
	}
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

func maxConsecutiveConsonants(s string) int {
	maxRun := 0
	currentRun := 0
	for _, r := range s {
		if unicode.IsLetter(r) && !isVowel(r) {
			currentRun++
			if currentRun > maxRun {
				maxRun = currentRun
			}
		} else {
			currentRun = 0
		}
	}
	return maxRun
}

func calculateVowelRatio(s string) float64 {
	vowels := 0
	letters := 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			letters++
			if isVowel(r) {
				vowels++
			}
		}
	}
	if letters == 0 {
		return 0.0
	}
	return float64(vowels) / float64(letters)
}

func calculateHexDensity(s string) float64 {
	if len(s) == 0 {
		return 0.0
	}
	hexCount := 0
	for _, r := range s {
		if isHexDigit(r) {
			hexCount++
		}
	}
	return float64(hexCount) / float64(len(s))
}

// Analyze evaluates a given domain and returns a threat assessment.
func (a *DGAAnalyzer) Analyze(domain string) ThreatResult {
	sld := ExtractSLD(domain)
	sldLower := strings.ToLower(sld)

	shannon := ShannonEntropy(sldLower)
	ngram := NGramEntropy(sldLower)
	maxCons := maxConsecutiveConsonants(sldLower)
	vowelRatio := calculateVowelRatio(sldLower)
	hexDensity := calculateHexDensity(sldLower)

	var indicators []string
	var score float64

	// 1. Shannon Entropy (>= 3.85 is severe algorithmic entropy)
	if shannon >= 3.85 {
		score += 0.45
		indicators = append(indicators, "Severe Shannon Entropy (>= 3.85)")
	} else if shannon >= 3.50 {
		score += 0.25
		indicators = append(indicators, "High Shannon Entropy (>= 3.50)")
	}

	// 2. Consecutive Consonants (>= 5 is unnatural)
	if maxCons >= 5 {
		score += 0.30
		indicators = append(indicators, "High Consonant Sequence (>= 5)")
	}

	// 3. Vowel Ratio (< 15% in length >= 8)
	if len(sldLower) >= 8 && vowelRatio < 0.15 {
		score += 0.20
		indicators = append(indicators, "Abnormally Low Vowel Ratio (< 15%)")
	}

	// 4. Hex / Base64 Hash Density
	if len(sldLower) >= 12 && hexDensity > 0.90 {
		score += 0.25
		indicators = append(indicators, "High Hexadecimal Density (> 90%)")
	}

	if score > 1.0 {
		score = 1.0
	}

	isThreat := score >= a.ThreatThreshold || shannon >= 3.85

	return ThreatResult{
		Domain:          domain,
		SLD:             sld,
		ShannonEntropy:  shannon,
		NGramEntropy:    ngram,
		ConsecutiveCons: maxCons,
		VowelRatio:      vowelRatio,
		HexDensity:      hexDensity,
		ThreatScore:     score,
		IsThreat:        isThreat,
		Indicators:      indicators,
	}
}
