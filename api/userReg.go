package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// UserRegistrationRequest is the structure for user registration
type UserRegistrationRequest struct {
	PhoneNumber  string `json:"phone_number"`
	Password     string `json:"password"`
	Name         string `json:"name"`
	ProfileImage string `json:"profile_image"`
	Address      string `json:"address"`
	GpsLocation  string `json:"gps_location"`
}

// RegisterUser handles user registration
func RegisterUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			http.Error(w, "Database connection not available", http.StatusInternalServerError)
			return
		}

		var req UserRegistrationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid data", http.StatusBadRequest)
			return
		}

		// Check for empty fields
		if req.Name == "" || req.PhoneNumber == "" || req.Password == "" {
			http.Error(w, "Fields cannot be empty", http.StatusBadRequest)
			return
		}

		// Ensure at least one of Address or GpsLocation is provided
		if req.Address == "" && req.GpsLocation == "" {
			http.Error(w, "Either address or GPS location must be provided", http.StatusBadRequest)
			return
		}

		// Trim whitespace
		req.PhoneNumber = trimSpace(req.PhoneNumber)
		req.Name = trimSpace(req.Name)
		req.Password = trimSpace(req.Password)
		req.Address = trimSpace(req.Address)
		req.GpsLocation = trimSpace(req.GpsLocation)

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

		// Insert new user into the database and retrieve the inserted ID
		result, err := db.Exec(
			"INSERT INTO Users (phone_number, password, name, profile_image, address, gps_location) VALUES (?, ?, ?, ?, ?, ST_GeomFromText(?))",
			req.PhoneNumber, hashedPassword, req.Name, req.ProfileImage, req.Address, req.GpsLocation,
		)
		if err != nil {
			log.Println("Error registering user:", err)
			http.Error(w, fmt.Sprintf("Error registering user: %v", err), http.StatusInternalServerError)
			return
		}

		// Get the ID of the inserted user
		userID, err := result.LastInsertId()
		if err != nil {
			log.Println("Error retrieving last insert ID:", err)
			http.Error(w, "Error retrieving user ID", http.StatusInternalServerError)
			return
		}

		// Send back success response with user ID
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "User registration successful",
			"id":      userID,
		})
	}
}
