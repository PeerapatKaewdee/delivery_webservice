package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

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

		// Check for empty fields
		if req.PhoneNumber == "" || req.Password == "" {
			http.Error(w, "Fields cannot be empty", http.StatusBadRequest)
			return
		}

		// Check if the user or rider exists and validate the password
		isRider, hashedPassword, err := getUserOrRiderPassword(db, req.PhoneNumber)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Compare the provided password with the stored hashed password
		if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Login successful
		if isRider {
			w.Write([]byte("Rider login successful"))
		} else {
			w.Write([]byte("User login successful"))
		}
	}
}

// getUserOrRiderPassword retrieves the hashed password for a given phone number
func getUserOrRiderPassword(db *sql.DB, phone string) (bool, string, error) {
	var hashedPassword string
	var isRider bool

	// Check in Riders table
	query := "SELECT password, 1 FROM Riders WHERE phone_number = ?"
	err := db.QueryRow(query, phone).Scan(&hashedPassword, &isRider)
	if err == nil {
		return true, hashedPassword, nil // Found in Riders
	}

	// Check in Users table
	query = "SELECT password, 0 FROM Users WHERE phone_number = ?"
	err = db.QueryRow(query, phone).Scan(&hashedPassword, &isRider)
	if err == nil {
		return false, hashedPassword, nil // Found in Users
	}

	return false, "", err // Not found
}
