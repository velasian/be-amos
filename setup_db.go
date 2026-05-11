package main

import (
	"fmt"
	"log"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Koneksi ke database default 'postgres' untuk bisa membuat database baru
	dsn := "host=127.0.0.1 user=postgres password=admin dbname=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal konek ke postgres default: %v", err)
	}

	// Mengeksekusi query pembuatan database
	err = db.Exec("CREATE DATABASE amos_db;").Error
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println("Database amos_db sudah ada, siap digunakan!")
		} else {
			log.Fatalf("Gagal membuat database amos_db: %v", err)
		}
	} else {
		fmt.Println("Database amos_db berhasil dibuat secara otomatis!")
	}
}
