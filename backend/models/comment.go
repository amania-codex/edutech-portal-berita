package models

import "gorm.io/gorm"

// Comment model — MVC Model layer
type Comment struct {
	gorm.Model
	NewsID         uint    `json:"news_id" gorm:"not null;index"`
	UserID         uint    `json:"user_id" gorm:"not null"`
	UserName       string  `json:"user_name" gorm:"type:varchar(100);not null"`
	Content        string  `json:"content" gorm:"type:text;not null"`
	Sentiment      string  `json:"sentiment" gorm:"type:varchar(20);default:'neutral'"`
	SentimentScore float64 `json:"sentiment_score" gorm:"default:0"`
	IsFlagged      bool    `json:"is_flagged" gorm:"default:false"`

	// Relations (for join queries)
	NewsTitle string `json:"news_title" gorm:"-"`
}
