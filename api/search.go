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
		// ตรวจสอบว่าเป็นคำขอ POST
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// สร้างตัวแปรสำหรับเก็บข้อมูลจาก Body
		var requestBody struct {
			Phone string `json:"phone"`
		}

		// ถอดรหัสข้อมูล JSON จาก Body
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Trim whitespace ของเบอร์โทรศัพท์
		phone := strings.TrimSpace(requestBody.Phone)

		// ตรวจสอบว่าใส่เบอร์โทรศัพท์หรือไม่
		if phone == "" {
			http.Error(w, "Phone number is required", http.StatusBadRequest)
			return
		}

		// ค้นหา user ที่ตรงกับเบอร์โทรศัพท์
		query := "SELECT uid, name FROM Users WHERE phone_number = ?"
		var receiverID int
		var receiverName string
		err = db.QueryRow(query, phone).Scan(&receiverID, &receiverName)
		if err == sql.ErrNoRows {
			http.Error(w, "Receiver not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Failed to search for receiver", http.StatusInternalServerError)
			return
		}

		// ส่งข้อมูลผู้รับกลับไปใน response
		response := map[string]interface{}{
			"receiver_id":   receiverID,
			"receiver_name": receiverName,
		}

		// กำหนด Content-Type ของ Header
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // ส่งสถานะ 200 OK
		json.NewEncoder(w).Encode(response)
	}
}
