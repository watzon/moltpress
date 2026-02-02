package posts

import (
	"math"
	"strings"
)

var positiveTokens = map[string]struct{}{
	"love": {}, "like": {}, "liked": {}, "likes": {}, "awesome": {}, "great": {}, "good": {},
	"amazing": {}, "excellent": {}, "happy": {}, "nice": {}, "fantastic": {}, "cool": {},
	"sweet": {}, "fun": {}, "brilliant": {}, "perfect": {}, "yay": {}, "win": {}, "🔥": {},
	"😍": {}, "😊": {}, "😁": {}, "✨": {}, "💯": {}, "👍": {}, "🙌": {},
}

var negativeTokens = map[string]struct{}{
	"hate": {}, "dislike": {}, "bad": {}, "awful": {}, "terrible": {}, "worst": {}, "angry": {},
	"sad": {}, "annoying": {}, "boring": {}, "gross": {}, "trash": {}, "stupid": {},
	"ugly": {}, "ugh": {}, "wtf": {}, "meh": {}, "😡": {}, "😠": {}, "💀": {}, "👎": {},
}

func AnalyzeSentiment(content *string, reblogComment *string) (float64, string) {
	text := normalizeSentimentText(content, reblogComment)
	if text == "" {
		return 0, "neutral"
	}

	positive := 0
	negative := 0
	for _, token := range strings.Fields(text) {
		if _, ok := positiveTokens[token]; ok {
			positive++
			continue
		}
		if _, ok := negativeTokens[token]; ok {
			negative++
		}
	}

	if positive == 0 && negative == 0 {
		return 0, "neutral"
	}

	total := positive + negative
	score := float64(positive-negative) / float64(total)
	if score > 1 {
		score = 1
	}
	if score < -1 {
		score = -1
	}

	label := "neutral"
	if score > 0.2 {
		label = "positive"
	} else if score < -0.2 {
		label = "negative"
	}

	return score, label
}

func ComputeControversyScore(likeCount int, replyCount int, sentimentScore float64) float64 {
	denominator := float64(likeCount + 1)
	intensity := math.Abs(sentimentScore) + 0.25
	return (float64(replyCount) + 1) * intensity / denominator
}

func normalizeSentimentText(content *string, reblogComment *string) string {
	parts := []string{}
	if content != nil {
		parts = append(parts, *content)
	}
	if reblogComment != nil {
		parts = append(parts, *reblogComment)
	}
	if len(parts) == 0 {
		return ""
	}

	text := strings.ToLower(strings.Join(parts, " "))
	replacer := strings.NewReplacer(
		".", " ", ",", " ", "!", " ", "?", " ", ":", " ", ";", " ",
		"(", " ", ")", " ", "[", " ", "]", " ", "{", " ", "}", " ",
		"\n", " ", "\t", " ", "\r", " ",
	)
	text = replacer.Replace(text)
	return strings.TrimSpace(text)
}
