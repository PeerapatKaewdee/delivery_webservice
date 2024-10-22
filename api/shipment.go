package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

// DeliveryRequest แสดงโครงสร้างข้อมูลการจัดส่ง
type DeliveryRequest struct {
	SenderID      int            `json:"sender_id"`
	ReceiverPhone string         `json:"receiver_phone,omitempty"` // Optional field
	Items         []ShipmentItem `json:"items"`
}

// ShipmentItem แสดงโครงสร้างข้อมูลสินค้าในการจัดส่ง
type ShipmentItem struct {
	Description string `json:"description"`
	Image       string `json:"image"`
}

// CreateDelivery สร้างรายการจัดส่งใหม่
func CreateDelivery(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req DeliveryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		// Trim the receiver phone if provided
		req.ReceiverPhone = strings.TrimSpace(req.ReceiverPhone)

		// Check if items are provided
		if len(req.Items) == 0 {
			http.Error(w, "Items cannot be empty", http.StatusBadRequest)
			return
		}

		var receiverID int
		if req.ReceiverPhone != "" {
			// Only check for receiver if phone is provided
			query := "SELECT uid FROM Users WHERE phone_number = ?"
			err := db.QueryRow(query, req.ReceiverPhone).Scan(&receiverID)
			if err != nil {
				http.Error(w, "Receiver not found", http.StatusNotFound)
				return
			}
		} else {
			receiverID = 0 // หรือใช้ค่า default อื่น ๆ ถ้าต้องการ
		}

		// เริ่มต้น Transaction
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
			return
		}

		// สร้าง Shipment
		insertQuery := "INSERT INTO Shipments (sender_id, receiver_id, status) VALUES (?, ?, ?)"
		result, err := tx.Exec(insertQuery, req.SenderID, receiverID, 1) // สถานะ 1: รอ Rider
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to create shipment", http.StatusInternalServerError)
			return
		}

		// รับ shipment_id ที่สร้างขึ้น
		shipmentID, err := result.LastInsertId()
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to retrieve shipment ID", http.StatusInternalServerError)
			return
		}

		// สร้าง Shipment Items
		for _, item := range req.Items {
			insertItemQuery := "INSERT INTO Shipment_Items (shipment_id, description, image) VALUES (?, ?, ?)"
			_, err := tx.Exec(insertItemQuery, shipmentID, item.Description, item.Image)
			if err != nil {
				tx.Rollback()
				http.Error(w, "Failed to create shipment item", http.StatusInternalServerError)
				return
			}
		}

		// ยืนยัน Transaction
		if err := tx.Commit(); err != nil {
			http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"message": "Delivery created successfully",
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}