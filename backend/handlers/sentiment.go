package handlers

import (
	"regexp"
	"strings"
)

// SentimentAnalysis — Business logic / sentiment & toxic comment filter
// Uses regex word boundaries to prevent substring false-positives.

var flaggedWords = []string{
	"anjing", "babi", "bangsat", "goblok", "idiot", "bodoh sekali", "kafir", "laknat",
	"bunuh", "hancur", "serang", "kebencian", "rasis", "diskriminasi",
	"fuck", "shit", "asshole", "bastard", "bitch",
}

var positiveWords = []string{
	"bagus", "baik", "hebat", "mantap", "keren", "luar biasa", "terima kasih", "sip", "setuju",
	"benar", "tepat", "informatif", "bermanfaat", "senang", "puas", "suka", "bangga",
	"inspiratif", "menarik", "berguna", "terbaik", "sempurna", "lanjutkan", "terimakasih",
	"memuaskan", "berkualitas", "profesional", "recommended", "wow", "amazing", "good", "great",
	"excellent", "helpful", "informative", "best", "love", "useful",
}

var negativeWords = []string{
	"buruk", "jelek", "bodoh", "tolol", "sampah", "tidak benar", "salah", "bohong", "tipu",
	"kecewa", "mengecewakan", "payah", "parah", "gagal", "menyesal", "menyedihkan", "benci",
	"tidak berguna", "hoax", "palsu", "fitnah", "kebohongan", "penipuan",
	"tidak setuju", "protes", "tidak suka", "rugi", "merugikan", "berbahaya",
	"bad", "terrible", "awful", "wrong", "fake", "useless", "worst",
}

var negations = []string{"tidak", "bukan", "kurang", "jangan", "tak", "gak", "nggak"}

type SentimentResult struct {
	Sentiment string  `json:"sentiment"`
	Score     float64 `json:"score"`
	IsFlagged bool    `json:"is_flagged"`
}

func buildWordPattern(words []string) *regexp.Regexp {
	var escaped []string
	for _, w := range words {
		escaped = append(escaped, regexp.QuoteMeta(w))
	}
	return regexp.MustCompile(`(?i)\b(` + strings.Join(escaped, "|") + `)\b`)
}

var (
	flaggedRegex  = buildWordPattern(flaggedWords)
	positiveRegex = buildWordPattern(positiveWords)
	negativeRegex = buildWordPattern(negativeWords)
	negationRegex = buildWordPattern(negations)
)

func AnalyzeSentiment(text string) SentimentResult {
	lower := strings.ToLower(text)

	// Check flagged inappropriate words
	isFlagged := flaggedRegex.MatchString(lower)

	// Tokenize words
	words := strings.Fields(regexp.MustCompile(`[^\w\s]`).ReplaceAllString(lower, " "))

	positiveScore := 0
	negativeScore := 0

	for i, w := range words {
		isNegated := false
		if i > 0 && negationRegex.MatchString(words[i-1]) {
			isNegated = true
		}

		if positiveRegex.MatchString(w) {
			if isNegated {
				negativeScore++ // e.g. "tidak bagus" -> negative
			} else {
				positiveScore++
			}
		} else if negativeRegex.MatchString(w) {
			if isNegated {
				positiveScore++ // e.g. "tidak buruk" -> positive
			} else {
				negativeScore++
			}
		}
	}

	total := positiveScore + negativeScore
	score := 0.0
	sentiment := "neutral"

	if total > 0 {
		score = float64(positiveScore-negativeScore) / float64(total)
		if score > 0.1 {
			sentiment = "positive"
		} else if score < -0.1 {
			sentiment = "negative"
		}
	}

	if isFlagged {
		sentiment = "negative"
		if score > -0.5 {
			score = -0.5
		}
	}

	score = float64(int(score*100)) / 100
	return SentimentResult{Sentiment: sentiment, Score: score, IsFlagged: isFlagged}
}