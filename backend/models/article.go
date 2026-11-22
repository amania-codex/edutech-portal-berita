package models

import "gorm.io/gorm"

// Article model — MVC Model layer
type Article struct {
	gorm.Model
	Title string `json:"title" gorm:"type:varchar(255);not null"`
	Slug  string `json:"slug" gorm:"type:varchar(255);unique;not null"`

	Excerpt      string `json:"excerpt" gorm:"type:text"`
	Content      string `json:"content" gorm:"type:longtext;not null"`
	ImageURL     string `json:"imageUrl" gorm:"type:varchar(500)"`
	Category     string `json:"category" gorm:"type:varchar(100);index"`
	CategorySlug string `json:"categorySlug" gorm:"type:varchar(150);index"`
	Author       string `json:"author" gorm:"type:varchar(100)"`
	AuthorID     uint   `json:"author_id" gorm:"default:0"`
	Tags         string `json:"tags" gorm:"type:varchar(500)"`
	ViewCount    int    `json:"viewCount" gorm:"default:0"`
	ReadTime     int    `json:"readTime" gorm:"default:5"`
	IsFeatured   bool   `json:"isFeatured" gorm:"default:false;index"`
	IsTrending   bool   `json:"isTrending" gorm:"default:false;index"`
	// Tanpa default:true — GORM mengabaikan nilai false pada kolom ber-default,
	// sehingga draft baru akan tersimpan sebagai terbit.
	IsPublished bool   `json:"isPublished" gorm:"not null;index"`
	Status      string `json:"status" gorm:"type:varchar(20);default:'published';index"`

	// WPID menyimpan ID post asal dari WordPress lama (0 = dibuat di sistem baru).
	// Dipakai agar impor bisa diulang tanpa duplikasi.
	WPID uint `json:"wpId" gorm:"column:wp_id;default:0;index"`

	// Slug semua kategori artikel (bukan kolom; diisi handler daftar berita)
	Categories []string `json:"categories" gorm:"-"`
}

// ArticleCategory — relasi many-to-many artikel ↔ kategori.
// Artikel hasil impor WordPress bisa berada di beberapa kategori sekaligus;
// kolom Article.Category tetap menyimpan kategori utama untuk tampilan badge.
type ArticleCategory struct {
	ArticleID  uint `gorm:"primaryKey;autoIncrement:false"`
	CategoryID uint `gorm:"primaryKey;autoIncrement:false;index"`
}

// ArticleRedirect — slug lama (mis. dari WordPress) yang harus diarahkan ke artikel.
type ArticleRedirect struct {
	ID        uint   `gorm:"primaryKey"`
	OldSlug   string `gorm:"type:varchar(255);uniqueIndex;not null"`
	ArticleID uint   `gorm:"not null;index"`
}
