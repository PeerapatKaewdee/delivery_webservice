package api

import (
	"database/sql"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// trimSpace ตัดช่องว่างนำหน้าและตามหลังจากสตริง
func trimSpace(s string) string {
	return strings.TrimSpace(s)
}

// phoneExists ตรวจสอบว่าหมายเลขโทรศัพท์มีอยู่ในตาราง Users หรือ Riders หรือไม่
func phoneExists(db *sql.DB, phone string) bool {
	var exists bool
	query := `
		SELECT EXISTS(SELECT 1 FROM Users WHERE phone_number = ?) 
		OR EXISTS(SELECT 1 FROM Riders WHERE phone_number = ?)`
	err := db.QueryRow(query, phone, phone).Scan(&exists)
	if err != nil {
		log.Printf("เกิดข้อผิดพลาดในการตรวจสอบหมายเลขโทรศัพท์: %v\n", err)
		return false
	}
	return exists
}

// hashPassword แฮชรหัสผ่านเป็นข้อความธรรมดาโดยใช้ bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
