package models

import "gorm.io/gorm"

type Notification struct {
	gorm.Model
	UserID    uint   `json:"user_id" gorm:"not null;index"`   // Penerima notifikasi
	SenderID  uint   `json:"sender_id" gorm:"not null"` // Pengirim (yg ngelike/reply)
	Type      string `json:"type" gorm:"type:varchar(20);not null"` // 'like', 'reply'
	Message   string `json:"message" gorm:"type:text;not null"`
	CommentID uint   `json:"comment_id" gorm:"default:0"`
	NewsSlug  string `json:"news_slug" gorm:"type:varchar(255)"` // URL referensi
	IsRead    bool   `json:"is_read" gorm:"default:false"`

	// Relasi virtual
	SenderName   string `json:"sender_name" gorm:"-"`
	SenderAvatar string `json:"sender_avatar" gorm:"-"`
}
