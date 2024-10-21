package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// RiderRegistrationRequest is the structure for rider registration
type RiderRegistrationRequest struct {
	PhoneNumber  string `json:"phone_number"`
	Password     string `json:"password"`
	Name         string `json:"name"`
	ProfileImage string `json:"profile_image"`
	LicensePlate string `json:"license_plate"`
}

// RegisterRider handles rider registration
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

		// Check for empty fields
		if req.Name == "" || req.PhoneNumber == "" || req.Password == "" || req.LicensePlate == "" {
			http.Error(w, "Fields cannot be empty", http.StatusBadRequest)
			return
		}

		// Trim whitespace
		req.PhoneNumber = trimSpace(req.PhoneNumber)
		req.Name = trimSpace(req.Name)
		req.Password = trimSpace(req.Password)
		req.LicensePlate = trimSpace(req.LicensePlate)

		// Check if phone number already exists
		if phoneExists(db, req.PhoneNumber) {
			http.Error(w, "Phone number already exists", http.StatusConflict)
			return
		}

		// Hash password
		hashedPassword, err := hashPassword(req.Password)
		if err != nil {
			log.Println("Error hashing password:", err)
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		// Insert new rider into the database and retrieve the inserted ID
		result, err := db.Exec(
			"INSERT INTO Riders (phone_number, password, name, profile_image, license_plate) VALUES (?, ?, ?, ?, ?)",
			req.PhoneNumber, hashedPassword, req.Name, req.ProfileImage, req.LicensePlate,
		)
		if err != nil {
			log.Println("Error registering rider:", err)
			http.Error(w, fmt.Sprintf("Error registering rider: %v", err), http.StatusInternalServerError)
			return
		}

		// Get the ID of the inserted rider
		riderID, err := result.LastInsertId()
		if err != nil {
			log.Println("Error retrieving last insert ID:", err)
			http.Error(w, "Error retrieving rider ID", http.StatusInternalServerError)
			return
		}

		// Send back success response with rider ID
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Rider registration successful",
			"id":      riderID,
		})
	}
}
