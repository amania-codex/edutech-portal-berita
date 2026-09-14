package handlers

import "strings"

// SentimentAnalysis â€” Port dari TypeScript sentimentAnalysis.ts
// MVC: Business logic / utility layer

var positiveWords = []string{
	"bagus", "baik", "hebat", "mantap", "keren", "luar biasa", "terima kasih", "sip", "setuju",
	"benar", "tepat", "informatif", "bermanfaat", "sangat", "senang", "puas", "suka", "bangga",
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

var flaggedWords = []string{
	"anjing", "babi", "bangsat", "goblok", "idiot", "bodoh sekali", "kafir", "laknat",
	"bunuh", "mati", "hancur", "serang", "kebencian", "rasis", "diskriminasi",
	"fuck", "shit", "ass", "bastard", "hate", "kill",
}

type SentimentResult struct {
	Sentiment string  `json:"sentiment"`
	Score     float64 `json:"score"`
	IsFlagged bool    `json:"is_flagged"`
}

func AnalyzeSentiment(text string) SentimentResult {
	lower := strings.ToLower(text)

	positiveScore := 0
	negativeScore := 0
	isFlagged := false

	for _, w := range flaggedWords {
		if strings.Contains(lower, w) {
			isFlagged = true
			break
		}
	}
	for _, w := range positiveWords {
		if strings.Contains(lower, w) {
			positiveScore++
		}
	}
	for _, w := range negativeWords {
		if strings.Contains(lower, w) {
			negativeScore++
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