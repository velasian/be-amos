package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword mengenkripsi string password menjadi hash bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash mengecek apakah password mentah cocok dengan hash di database
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
