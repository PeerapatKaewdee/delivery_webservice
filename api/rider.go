package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// RiderRegistrationRequest คือโครงสร้างข้อมูลสำหรับการลงทะเบียนผู้ขับขี่
type RiderRegistrationRequest struct {
	Rid          int    `json:"rid"`
	PhoneNumber  string `json:"phone_number"`
	Password     string `json:"password"`
	Name         string `json:"name"`
	ProfileImage string `json:"profile_image"`
	LicensePlate string `json:"license_plate"`
}

// RegisterRider จัดการการลงทะเบียนผู้ขับขี่
func RegisterRider(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			http.Error(w, "Database connection not available", http.StatusInternalServerError)
			return
		}

		var req RiderRegistrationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid data", http.StatusBadRequest)
			return
		}

		// ตรวจสอบฟิลด์ที่ว่าง
		if req.Name == "" || req.PhoneNumber == "" || req.Password == "" || req.LicensePlate == "" {
			http.Error(w, "Fields cannot be empty", http.StatusBadRequest)
			return
		}

		// ตัดช่องว่าง
		req.PhoneNumber = trimSpace(req.PhoneNumber)
		req.Name = trimSpace(req.Name)
		req.Password = trimSpace(req.Password)
		req.LicensePlate = trimSpace(req.LicensePlate)

		// ตรวจสอบหมายเลขโทรศัพท์
		if phoneExists(db, req.PhoneNumber) {
			http.Error(w, "Phone number already exists", http.StatusConflict)
			return
		}

		// แฮชรหัสผ่าน
		hashedPassword, err := hashPassword(req.Password)
		if err != nil {
			log.Println("Error hashing password:", err)
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		// แทรกผู้ขับขี่ใหม่ลงในฐานข้อมูล
result, err := db.Exec(
	"INSERT INTO Riders (phone_number, password, name, profile_image, license_plate) VALUES (?, ?, ?, ?, ?)",
	req.PhoneNumber, hashedPassword, req.Name, req.ProfileImage, req.LicensePlate,
)

// ตรวจสอบข้อผิดพลาด
if err != nil {
	log.Println("Error registering rider:", err)
	http.Error(w, fmt.Sprintf("Error registering rider: %v", err), http.StatusInternalServerError)
	return
}

// ดึง ID ของผู้ขับขี่ที่ถูกสร้างขึ้นมาใหม่ (ถ้าตารางมีการกำหนด auto-increment)
riderID, err := result.LastInsertId()
if err != nil {
	log.Println("Error retrieving last insert ID:", err)
	http.Error(w, fmt.Sprintf("Error retrieving last insert ID: %v", err), http.StatusInternalServerError)
	return
}

// ส่งกลับข้อความยืนยันพร้อม ID ของผู้ขับขี่
w.WriteHeader(http.StatusCreated)
json.NewEncoder(w).Encode(map[string]interface{}{
	"message":   "Rider registration successful",
	"rider_id":  riderID, // ส่งกลับ ID ของผู้ขับขี่ใหม่
})

	}
}


// RiderLicensePlateResponse is the structure for the response containing the rider's license plate
type RiderLicensePlateResponse struct {
	LicensePlate string `json:"license_plate"`
}

// GetRider handles fetching rider's license plate information
func GetRider(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			http.Error(w, "Database connection not available", http.StatusInternalServerError)
			return
		}

		// Extract rider ID from the URL path (assuming the URL is like /get/rider/{rider_id})
		vars := mux.Vars(r) // หากใช้ Gorilla Mux
		riderID := vars["rider_id"]

		// Query the rider's license plate from the database
		var response RiderLicensePlateResponse
		err := db.QueryRow("SELECT license_plate FROM Riders WHERE rid = ?", riderID).Scan(
			&response.LicensePlate,
		)

		if err == sql.ErrNoRows {
			http.Error(w, "Rider not found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Println("Error fetching rider:", err)
			http.Error(w, fmt.Sprintf("Error fetching rider: %v", err), http.StatusInternalServerError)
			return
		}

		// Send back the rider's license plate information as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
