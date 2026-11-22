package database

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"news-portal-backend/models"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Arsip berita terbit (dibuat oleh tools/export-seed.php).
//
//go:embed seed/berita.json.gz
var newsSeed []byte

type seedCategory struct {
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Parent string `json:"parent"`
	WPID   uint   `json:"wpId"`
}

type seedArticle struct {
	WPID         uint     `json:"wpId"`
	CreatedAt    string   `json:"createdAt"`
	UpdatedAt    string   `json:"updatedAt"`
	Title        string   `json:"title"`
	Slug         string   `json:"slug"`
	Excerpt      string   `json:"excerpt"`
	Content      string   `json:"content"`
	ImageURL     string   `json:"imageUrl"`
	Category     string   `json:"category"`
	CategorySlug string   `json:"categorySlug"`
	Author       string   `json:"author"`
	Tags         string   `json:"tags"`
	ViewCount    int      `json:"viewCount"`
	ReadTime     int      `json:"readTime"`
	IsFeatured   bool     `json:"isFeatured"`
	Categories   []string `json:"categories"`
	OldSlugs     []string `json:"oldSlugs"`
}

// SeedNews mengisi arsip berita bila database belum pernah memuatnya.
//
// Hanya berjalan sekali: bila sudah ada satu saja artikel arsip (termasuk yang dihapus admin),
// tidak ada yang disentuh. Matikan dengan SEED_NEWS=false.
func SeedNews() {
	if os.Getenv("SEED_NEWS") == "false" {
		return
	}
	var existing int64
	if err := DB.Unscoped().Model(&models.Article{}).Where("wp_id > 0").Count(&existing).Error; err != nil {
		log.Println("SeedNews: gagal memeriksa artikel:", err)
		return
	}
	if existing > 0 {
		return
	}
	start := time.Now()
	n, err := loadNewsSeed()
	if err != nil {
		// transaksi dibatalkan: tidak ada data setengah jadi, dicoba lagi saat backend dinyalakan ulang
		log.Println("SeedNews: gagal, semua perubahan dibatalkan:", err)
		return
	}
	log.Printf("SeedNews: %d berita arsip dimuat (%s)", n, time.Since(start).Round(time.Millisecond))
}

func loadNewsSeed() (int, error) {
	zr, err := gzip.NewReader(bytes.NewReader(newsSeed))
	if err != nil {
		return 0, err
	}
	defer zr.Close()
	var seed struct {
		Categories []seedCategory `json:"categories"`
		Articles   []seedArticle  `json:"articles"`
	}
	if err := json.NewDecoder(zr).Decode(&seed); err != nil {
		return 0, err
	}

	// Waktu pada arsip = WIB
	wib := time.FixedZone("WIB", 7*60*60)
	parse := func(s string) (time.Time, error) { return time.ParseInLocation("2006-01-02 15:04:05", s, wib) }

	total := 0
	quiet := DB.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
	err = quiet.Transaction(func(tx *gorm.DB) error {
		// Kategori: pakai yang sudah ada (slug/nama sama), buat sisanya; induk selalu lebih dulu
		catID := map[string]uint{}
		for _, c := range seed.Categories {
			var cat models.Category
			if err := tx.Where("slug = ? OR name = ?", c.Slug, c.Name).Limit(1).Find(&cat).Error; err != nil {
				return err
			}
			if cat.ID == 0 {
				cat = models.Category{Name: c.Name, Slug: c.Slug, WPID: c.WPID}
				if parent, ok := catID[c.Parent]; ok && c.Parent != "" {
					cat.ParentID = &parent
				}
				if err := tx.Create(&cat).Error; err != nil {
					return fmt.Errorf("kategori %q: %w", c.Name, err)
				}
			}
			catID[c.Slug] = cat.ID
		}

		// Slug yang sudah terpakai di database ini (termasuk artikel terhapus)
		var used []string
		if err := tx.Unscoped().Model(&models.Article{}).Pluck("slug", &used).Error; err != nil {
			return err
		}
		taken := make(map[string]bool, len(used))
		for _, s := range used {
			taken[s] = true
		}

		articles := make([]models.Article, 0, len(seed.Articles))
		sources := make([]seedArticle, 0, len(seed.Articles))
		for _, a := range seed.Articles {
			if taken[a.Slug] {
				if a.WPID == 0 {
					continue // artikel yang sama sudah ada
				}
				a.Slug = fmt.Sprintf("%s-%d", a.Slug, a.WPID)
			}
			taken[a.Slug] = true
			created, err := parse(a.CreatedAt)
			if err != nil {
				return fmt.Errorf("tanggal %q: %w", a.Slug, err)
			}
			updated, err := parse(a.UpdatedAt)
			if err != nil {
				updated = created
			}
			art := models.Article{
				Title: a.Title, Slug: a.Slug, Excerpt: a.Excerpt, Content: a.Content, ImageURL: a.ImageURL,
				Category: a.Category, CategorySlug: a.CategorySlug, Author: a.Author, Tags: a.Tags,
				ViewCount: a.ViewCount, ReadTime: a.ReadTime, IsFeatured: a.IsFeatured,
				IsPublished: true, Status: "published", WPID: a.WPID,
			}
			art.CreatedAt, art.UpdatedAt = created, updated
			articles = append(articles, art)
			sources = append(sources, a)
		}
		if err := tx.CreateInBatches(&articles, 100).Error; err != nil {
			return err
		}

		var links []models.ArticleCategory
		var redirects []models.ArticleRedirect
		for i, a := range sources {
			seen := map[uint]bool{}
			for _, slug := range a.Categories {
				if id := catID[slug]; id != 0 && !seen[id] {
					seen[id] = true
					links = append(links, models.ArticleCategory{ArticleID: articles[i].ID, CategoryID: id})
				}
			}
			for _, old := range a.OldSlugs {
				redirects = append(redirects, models.ArticleRedirect{OldSlug: old, ArticleID: articles[i].ID})
			}
		}
		if len(links) > 0 {
			if err := tx.CreateInBatches(&links, 500).Error; err != nil {
				return err
			}
		}
		if len(redirects) > 0 {
			if err := tx.CreateInBatches(&redirects, 500).Error; err != nil {
				return err
			}
		}
		total = len(articles)
		return nil
	})
	return total, err
}
