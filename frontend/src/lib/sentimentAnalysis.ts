/**
 * SENTIMENT ANALYSIS MODULE
 * Analyzes comment sentiment using keyword-based approach.
 * Architecture: MVC - Business logic utility layer.
 */

interface SentimentResult {
  sentiment: 'positive' | 'negative' | 'neutral';
  score: number; // -1 to 1
  isFlagged: boolean;
  keywords: string[];
}

const POSITIVE_WORDS = [
  'bagus', 'baik', 'hebat', 'mantap', 'keren', 'luar biasa', 'terima kasih', 'sip', 'setuju',
  'benar', 'tepat', 'informatif', 'bermanfaat', 'sangat', 'senang', 'puas', 'suka', 'bangga',
  'inspiratif', 'menarik', 'berguna', 'terbaik', 'sempurna', 'lanjutkan', 'terimakasih',
  'memuaskan', 'berkualitas', 'profesional', 'recommended', 'wow', 'amazing', 'good', 'great',
  'excellent', 'helpful', 'informative', 'best', 'love', 'useful',
];

const NEGATIVE_WORDS = [
  'buruk', 'jelek', 'bodoh', 'tolol', 'sampah', 'tidak benar', 'salah', 'bohong', 'tipu',
  'kecewa', 'mengecewakan', 'payah', 'parah', 'gagal', 'menyesal', 'menyedihkan', 'benci',
  'tidak berguna', 'hoax', 'palsu', 'fitnah', 'fitnah', 'kebohongan', 'penipuan',
  'tidak setuju', 'protes', 'tidak suka', 'rugi', 'merugikan', 'berbahaya',
  'bad', 'terrible', 'awful', 'wrong', 'fake', 'useless', 'worst',
];

const FLAGGED_WORDS = [
  'anjing', 'babi', 'bangsat', 'goblok', 'idiot', 'bodoh sekali', 'kafir', 'laknat',
  'bunuh', 'mati', 'hancur', 'serang', 'kebencian', 'rasis', 'diskriminasi',
  'fuck', 'shit', 'ass', 'bastard', 'hate', 'kill',
];

export function analyzeSentiment(text: string): SentimentResult {
  const lowerText = text.toLowerCase();
  const words = lowerText.split(/\s+/);
  
  let positiveScore = 0;
  let negativeScore = 0;
  const foundKeywords: string[] = [];
  let isFlagged = false;

  // Check for flagged content first
  for (const flagWord of FLAGGED_WORDS) {
    if (lowerText.includes(flagWord)) {
      isFlagged = true;
      foundKeywords.push(flagWord);
    }
  }

  // Count positive words
  for (const posWord of POSITIVE_WORDS) {
    if (lowerText.includes(posWord)) {
      positiveScore++;
      foundKeywords.push(posWord);
    }
  }

  // Count negative words
  for (const negWord of NEGATIVE_WORDS) {
    if (lowerText.includes(negWord)) {
      negativeScore++;
      foundKeywords.push(negWord);
    }
  }

  // Calculate score (-1 to 1)
  const total = positiveScore + negativeScore;
  let score = 0;
  let sentiment: 'positive' | 'negative' | 'neutral' = 'neutral';

  if (total > 0) {
    score = (positiveScore - negativeScore) / total;
    if (score > 0.1) sentiment = 'positive';
    else if (score < -0.1) sentiment = 'negative';
  }

  // If flagged, override to negative
  if (isFlagged) {
    sentiment = 'negative';
    score = Math.min(score, -0.5);
  }

  return {
    sentiment,
    score: Math.round(score * 100) / 100,
    isFlagged,
    keywords: [...new Set(foundKeywords)],
  };
}

export function getSentimentLabel(sentiment: string): string {
  switch (sentiment) {
    case 'positive': return 'Positif';
    case 'negative': return 'Negatif';
    default: return 'Netral';
  }
}

export function getSentimentColor(sentiment: string): string {
  switch (sentiment) {
    case 'positive': return '#22c55e';
    case 'negative': return '#ef4444';
    default: return '#94a3b8';
  }
}

export function getSentimentEmoji(sentiment: string): string {
  switch (sentiment) {
    case 'positive': return '😊';
    case 'negative': return '😠';
    default: return '😐';
  }
}
