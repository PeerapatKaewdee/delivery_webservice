package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// RiderRegistrationRequest เป็นโครงสร้างสำหรับการลงทะเบียนไรเดอร์
type RiderRegistrationRequest struct {
	PhoneNumber  string `json:"phone_number"`
	Password     string `json:"password"`
	Name         string `json:"name"`
	ProfileImage string `json:"profile_image"`
	LicensePlate string `json:"license_plate"`
}

// RegisterRider ฟังก์ชันสำหรับจัดการการลงทะเบียนไรเดอร์
func RegisterRider(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			http.Error(w, "การเชื่อมต่อฐานข้อมูลไม่พร้อมใช้งาน", http.StatusInternalServerError)
			return
		}

		var req RiderRegistrationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "ข้อมูลไม่ถูกต้อง", http.StatusBadRequest)
			return
		}

		// ตรวจสอบฟิลด์ที่ว่างเปล่า
		if req.Name == "" || req.PhoneNumber == "" || req.Password == "" || req.LicensePlate == "" {
			http.Error(w, "ฟิลด์ไม่สามารถว่างเปล่าได้", http.StatusBadRequest)
			return
		}

		// ตัดช่องว่าง
		req.PhoneNumber = trimSpace(req.PhoneNumber)
		req.Name = trimSpace(req.Name)
		req.Password = trimSpace(req.Password)
		req.LicensePlate = trimSpace(req.LicensePlate)

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

		// แทรกไรเดอร์ใหม่ลงในฐานข้อมูล
		_, err = db.Exec(
			"INSERT INTO Riders (phone_number, password, name, profile_image, license_plate) VALUES (?, ?, ?, ?, ?)",
			req.PhoneNumber, hashedPassword, req.Name, req.ProfileImage, req.LicensePlate,
		)
		if err != nil {
			log.Println("เกิดข้อผิดพลาดในการลงทะเบียนไรเดอร์:", err)
			http.Error(w, fmt.Sprintf("เกิดข้อผิดพลาดในการลงทะเบียนไรเดอร์: %v", err), http.StatusInternalServerError)
			return
		}

		// ส่งกลับคำตอบสำเร็จ
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "ลงทะเบียนไรเดอร์สำเร็จ"})
	}
}
