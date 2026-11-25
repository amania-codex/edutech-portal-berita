package models

import "time"

// ActivityLog — catatan siapa melakukan apa (berita, pengguna, video).
// Tidak bisa diubah atau dihapus dari aplikasi, supaya bisa dipakai sebagai bukti.
type ActivityLog struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time `json:"createdAt" gorm:"index"`
	UserID      uint      `json:"userId" gorm:"index"`
	UserName    string    `json:"userName" gorm:"type:varchar(100)"`
	UserRole    string    `json:"userRole" gorm:"type:varchar(20)"`
	Action      string    `json:"action" gorm:"type:varchar(30);index"`
	TargetType  string    `json:"targetType" gorm:"type:varchar(20);index:idx_activity_target"`
	TargetID    uint      `json:"targetId" gorm:"index:idx_activity_target"`
	TargetTitle string    `json:"targetTitle" gorm:"type:varchar(255)"`
	Detail      string    `json:"detail" gorm:"type:text"`
	IP          string    `json:"ip" gorm:"type:varchar(64)"`
}
