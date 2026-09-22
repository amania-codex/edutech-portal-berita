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

	ParentID       *uint   `json:"parent_id" gorm:"index"` // Untuk fitur reply
	LikeCount      int     `json:"like_count" gorm:"default:0"`

	// Relations (for join queries)
	NewsTitle string `json:"news_title" gorm:"-"`
	Replies   []Comment `json:"replies" gorm:"foreignKey:ParentID"`
	IsLiked   bool      `json:"is_liked" gorm:"-"`
}

type CommentLike struct {
	gorm.Model
	CommentID uint `json:"comment_id" gorm:"not null;index"`
	UserID    uint `json:"user_id" gorm:"not null;index"`
}
