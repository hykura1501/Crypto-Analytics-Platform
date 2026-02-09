package sentiment

import (
	"math"
	"strings"
	"unicode"
)

// Analyzer performs sentiment analysis
type Analyzer struct {
	positiveWords map[string]bool
	negativeWords map[string]bool
}

// NewAnalyzer creates a new sentiment analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		positiveWords: loadPositiveWords(),
		negativeWords: loadNegativeWords(),
	}
}

// Result contains sentiment analysis results
type Result struct {
	Compound float64
	Pos      float64
	Neg      float64
	Neu      float64
	Label    string
}

// Analyze performs sentiment analysis on text
func (a *Analyzer) Analyze(text string) Result {
	words := tokenize(text)
	
	posScore := 0.0
	negScore := 0.0
	neutralCount := 0
	
	for _, word := range words {
		wordLower := strings.ToLower(word)
		
		if a.positiveWords[wordLower] {
			posScore++
		} else if a.negativeWords[wordLower] {
			negScore++
		} else {
			neutralCount++
		}
	}
	
	total := float64(len(words))
	if total == 0 {
		return Result{
			Compound: 0.0,
			Pos:      0.0,
			Neg:      0.0,
			Neu:      1.0,
			Label:    "Neutral",
		}
	}
	
	pos := posScore / total
	neg := negScore / total
	neu := float64(neutralCount) / total
	
	// Calculate compound score (normalized difference)
	compound := (posScore - negScore) / math.Max(total, 1.0)
	compound = math.Max(-1.0, math.Min(1.0, compound))
	
	label := "Neutral"
	if compound >= 0.05 {
		label = "Positive"
	} else if compound <= -0.05 {
		label = "Negative"
	}
	
	return Result{
		Compound: compound,
		Pos:      pos,
		Neg:      neg,
		Neu:      neu,
		Label:    label,
	}
}

func tokenize(text string) []string {
	words := []string{}
	current := strings.Builder{}
	
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(unicode.ToLower(r))
		} else {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		}
	}
	
	if current.Len() > 0 {
		words = append(words, current.String())
	}
	
	return words
}

func loadPositiveWords() map[string]bool {
	words := []string{
		"good", "great", "excellent", "amazing", "wonderful", "fantastic",
		"positive", "bullish", "rise", "surge", "gain", "profit", "success",
		"up", "high", "strong", "growth", "increase", "boost", "rally",
		"breakthrough", "innovation", "adoption", "partnership", "launch",
		"approval", "support", "investment", "funding", "milestone",
		"tăng", "tốt", "tích cực", "tăng trưởng", "phát triển", "thành công",
	}
	
	result := make(map[string]bool)
	for _, w := range words {
		result[w] = true
	}
	return result
}

func loadNegativeWords() map[string]bool {
	words := []string{
		"bad", "terrible", "awful", "horrible", "negative", "bearish",
		"fall", "drop", "crash", "loss", "decline", "decrease", "fail",
		"down", "low", "weak", "risk", "concern", "worry", "fear",
		"hack", "attack", "scam", "fraud", "ban", "regulation", "ban",
		"lawsuit", "fine", "penalty", "rejection", "delay", "problem",
		"giảm", "xấu", "tiêu cực", "sụt giảm", "thất bại", "rủi ro",
	}
	
	result := make(map[string]bool)
	for _, w := range words {
		result[w] = true
	}
	return result
}

