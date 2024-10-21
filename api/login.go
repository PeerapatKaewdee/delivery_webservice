package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// LoginRequest represents the login request structure
type LoginRequest struct {
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

// LoginUserOrRider handles login for both users and riders
func LoginUserOrRider(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		// Trim spaces and check for empty fields
		req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)
		req.Password = strings.TrimSpace(req.Password)
		if req.PhoneNumber == "" || req.Password == "" {
			http.Error(w, "Phone number and password cannot be empty", http.StatusBadRequest)
			return
		}

		// Check if the user or rider exists and validate the password
		isRider, id, hashedPassword, err := getUserOrRiderDetails(db, req.PhoneNumber)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Compare the provided password with the stored hashed password
		if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
			log.Printf("Password mismatch. DB Hashed Password: %s, Entered Password: %s", hashedPassword, req.Password)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Determine the user type ("rider" or "user")
		userType := "user"
		if isRider {
			userType = "rider"
		}

		// Sending back ID and type information in response
		response := map[string]interface{}{
			"id":   id,       // Either "uid" or "rid"
			"type": userType, // Either "rider" or "user"
		}

		json.NewEncoder(w).Encode(response)
	}
}

// getUserOrRiderDetails retrieves the correct ID, hashed password, and user type for a given phone number
func getUserOrRiderDetails(db *sql.DB, phone string) (bool, int, string, error) {
	var hashedPassword string
	var isRider bool
	var id int

	log.Printf("Looking for phone number: %s", phone) // Log phone number

	// Check in Riders table
	query := "SELECT rid, password, 1 FROM Riders WHERE phone_number = ?"
	err := db.QueryRow(query, phone).Scan(&id, &hashedPassword, &isRider)
	if err == nil {
		log.Printf("Found Rider with ID %d", id) // Log found Rider
		return true, id, hashedPassword, nil // Found in Riders
	}

	// Check in Users table
	query = "SELECT uid, password, 0 FROM Users WHERE phone_number = ?"
	err = db.QueryRow(query, phone).Scan(&id, &hashedPassword, &isRider)
	if err == nil {
		log.Printf("Found User with ID %d", id) // Log found User
		return false, id, hashedPassword, nil // Found in Users
	}

	log.Printf("No match for phone number: %s", phone) // Log no match
	return false, 0, "", err // Not found
}
