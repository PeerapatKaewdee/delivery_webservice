package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// UserRegistrationRequest เป็นโครงสร้างสำหรับการลงทะเบียนผู้ใช้
type UserRegistrationRequest struct {
	PhoneNumber  string `json:"phone_number"`
	Password     string `json:"password"`
	Name         string `json:"name"`
	ProfileImage string `json:"profile_image"`
	Address      string `json:"address"`
	GpsLocation  string `json:"gps_location"`
}

// RegisterUser ฟังก์ชันสำหรับจัดการการลงทะเบียนผู้ใช้
func RegisterUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			http.Error(w, "การเชื่อมต่อฐานข้อมูลไม่พร้อมใช้งาน", http.StatusInternalServerError)
			return
		}

		var req UserRegistrationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "ข้อมูลไม่ถูกต้อง", http.StatusBadRequest)
			return
		}

		// ตรวจสอบฟิลด์ที่ว่างเปล่า
		if req.Name == "" || req.PhoneNumber == "" || req.Password == "" {
			http.Error(w, "ฟิลด์ไม่สามารถว่างเปล่าได้", http.StatusBadRequest)
			return
		}

		// ตัดช่องว่าง
		req.PhoneNumber = trimSpace(req.PhoneNumber)
		req.Name = trimSpace(req.Name)
		req.Password = trimSpace(req.Password)
		req.Address = trimSpace(req.Address)
		req.GpsLocation = trimSpace(req.GpsLocation)

		// ตรวจสอบหมายเลขโทรศัพท์ที่มีอยู่แล้ว
		if phoneExists(db, req.PhoneNumber) {
			http.Error(w, "หมายเลขโทรศัพท์มีอยู่แล้ว", http.StatusConflict)
			return
		}

		// แฮชรหัสผ่าน
		hashedPassword, err := hashPassword(req.Password)
		if err != nil {
			log.Println("เกิดข้อผิดพลาดในการแฮชรหัสผ่าน:", err)
			http.Error(w, "เกิดข้อผิดพลาดในการแฮชรหัสผ่าน", http.StatusInternalServerError)
			return
		}

		// แทรกผู้ใช้ใหม่ลงในฐานข้อมูล
		_, err = db.Exec(
			"INSERT INTO Users (phone_number, password, name, profile_image, address, gps_location) VALUES (?, ?, ?, ?, ?, ST_GeomFromText(?))",
			req.PhoneNumber, hashedPassword, req.Name, req.ProfileImage, req.Address, req.GpsLocation,
		)
		if err != nil {
			log.Println("เกิดข้อผิดพลาดในการลงทะเบียนผู้ใช้:", err)
			http.Error(w, fmt.Sprintf("เกิดข้อผิดพลาดในการลงทะเบียนผู้ใช้: %v", err), http.StatusInternalServerError)
			return
		}

		// ส่งกลับคำตอบสำเร็จ
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "ลงทะเบียนผู้ใช้สำเร็จ"})
	}
}
