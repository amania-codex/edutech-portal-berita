package database

import (
	"log"

	"news-portal-backend/models"
)

// seedMenuDefaults mengisi menu navigasi dan Topik Pilihan bawaan (sekali, saat kolomnya baru dibuat).
// Setelah itu sepenuhnya diatur dari Dashboard > Kelola Kategori.
func seedMenuDefaults() {
	menu := []string{"berita", "pendidikan", "perguruan-tinggi", "jateng-diy", "nasional", "pariwisata", "artikel"}
	for i, slug := range menu {
		DB.Model(&models.Category{}).Where("slug = ?", slug).Update("menu_order", i+1)
	}
	DB.Model(&models.Category{}).Where("slug IN ?", []string{"perguruan-tinggi", "jateng-diy", "pariwisata"}).Update("is_topic", true)
	log.Println("Menu kategori bawaan diisi.")
}
