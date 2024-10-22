package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

// SearchReceiverByPhone ค้นหาผู้รับตามเบอร์โทรศัพท์
func SearchReceiverByPhone(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var requestBody struct {
			Phone   string `json:"phone"`
			UserID  int    `json:"user_id"` // รับ ID ของผู้ใช้ที่ล็อกอิน
		}

		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		phone := strings.TrimSpace(requestBody.Phone)
		if phone == "" {
			http.Error(w, "Phone number is required", http.StatusBadRequest)
			return
		}

		query := "SELECT uid, name, phone_number FROM Users WHERE phone_number LIKE ?"
		rows, err := db.Query(query, phone+"%")
		if err != nil {
			http.Error(w, "Failed to search for receiver", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var users []map[string]interface{}
		for rows.Next() {
			var receiverID int
			var receiverName, receiverPhone string
			if err := rows.Scan(&receiverID, &receiverName, &receiverPhone); err != nil {
				http.Error(w, "Error scanning row", http.StatusInternalServerError)
				return
			}
			// ตรวจสอบว่าผู้รับเป็นผู้ใช้ที่ล็อกอินหรือไม่
			if receiverID != requestBody.UserID {
				users = append(users, map[string]interface{}{
					"receiver_id":   receiverID,
					"receiver_name": receiverName,
					"receiver_phone": receiverPhone,
				})
			}
		}

		if len(users) == 0 {
			http.Error(w, "Receiver not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(users)
	}
}