package api

import (
	"database/sql"
	"encoding/json"
	"fmt" // เพิ่ม fmt สำหรับการ Debug
	"net/http"
	"strings"

	"delivery_webservice/model"

	"golang.org/x/crypto/bcrypt"
)

// RegisterRider handles rider registration
func RegisterRider(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req model.RiderRegistrationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		// Check for empty fields
		if req.Name == "" || req.PhoneNumber == "" || req.LicensePlate == "" || req.Password == "" {
			http.Error(w, "Fields cannot be empty", http.StatusBadRequest)
			return
		}

		// Trim spaces
		req.PhoneNumber = trimSpace(req.PhoneNumber)
		req.Name = trimSpace(req.Name)
		req.LicensePlate = trimSpace(req.LicensePlate)
		req.Password = trimSpace(req.Password)

		// Check if phone number already exists in Users or Riders table
		if phoneExists(db, req.PhoneNumber) {
			http.Error(w, "Phone number already exists", http.StatusConflict)
			return
		}

		// Hash the password
		hashedPassword, err := hashPassword(req.Password)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		// Insert the new rider into the database
		_, err = db.Exec(
			"INSERT INTO Riders (phone_number, password, name, profile_image, license_plate) VALUES (?, ?, ?, ?, ?)",
			req.PhoneNumber, hashedPassword, req.Name, req.ProfileImage, req.LicensePlate,
		)
		if err != nil {
			// แสดงรายละเอียดของ error
			http.Error(w, fmt.Sprintf("Error registering rider: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Rider registered successfully"))
	}
}

// phoneExists checks if the phone number exists in either Users or Riders table
func phoneExists(db *sql.DB, phone string) bool {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM Users WHERE phone_number = ?
			UNION
			SELECT 1 FROM Riders WHERE phone_number = ?
		)
	`
	err := db.QueryRow(query, phone, phone).Scan(&exists)
	if err != nil {
		fmt.Printf("Error checking phone number: %v\n", err) // แสดง error ใน console สำหรับการ Debug
		return false
	}
	return exists
}

// hashPassword hashes a plaintext password using bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// trimSpace removes leading and trailing spaces from a string
func trimSpace(s string) string {
	return strings.TrimSpace(s)
}
