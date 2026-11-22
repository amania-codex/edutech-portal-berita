package main

import (
	"log"
	"os"
	"strings"
)

// loadDotEnv membaca file .env (KUNCI=nilai per baris) ke environment.
// Variabel yang sudah ada di environment (systemd, pm2, shell) tidak ditimpa.
func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return // tanpa .env: pakai environment apa adanya
	}
	values := map[string]string{}
	for _, line := range strings.Split(strings.TrimPrefix(string(data), string(rune(0xFEFF))), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(strings.TrimPrefix(line, "export "), "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !ok || key == "" {
			continue
		}
		if n := len(value); n >= 2 && (value[0] == '"' || value[0] == '\'') && value[n-1] == value[0] {
			value = value[1 : n-1]
		}
		values[key] = value // bila dobel, baris terakhir yang dipakai
	}
	for key, value := range values {
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
	log.Printf("Konfigurasi dibaca dari %s", path)
}

// listenAddr — HOST dan PORT dari environment (bawaan: semua alamat, port 3000).
func listenAddr() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	return os.Getenv("HOST") + ":" + port
}
