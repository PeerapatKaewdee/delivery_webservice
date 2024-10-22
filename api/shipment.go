package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// ShipmentItem แสดงโครงสร้างข้อมูลสินค้าในการจัดส่ง
//
//	type ShipmentItem struct {
//		IID         int    `json:"iid"`
//		Description string `json:"description"`
//		Image       string `json:"image"`
//	}
type ShipmentItem struct {
	IID         int            `json:"iid"`
	Description string         `json:"description"`
	Image       sql.NullString `json:"image"` // ใช้ sql.NullString
}

// DeliveryRequest แสดงโครงสร้างข้อมูลการจัดส่ง
type DeliveryRequest struct {
	SenderID      int            `json:"sender_id"`
	ReceiverPhone string         `json:"receiver_phone,omitempty"` // Optional field
	Items         []ShipmentItem `json:"items"`
}

type ShipmentDetail struct {
	ShipmentID int            `json:"shipment_id"`
	SenderID   string         `json:"sender_id"`
	ReceiverID string         `json:"receiver_id"`
	RiderID    string         `json:"rider_id"`
	Status     string         `json:"status"`
	Items      []ShipmentItem `json:"items"`
}

// type Shipment_id struct {
// 	IID         int    `json:"iid"`         // ID ของสินค้า
// 	Description string `json:"description"` // คำอธิบายของสินค้า
// 	Image       string `json:"image"`       // URL ของภาพสินค้า
// }

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

// GetDeliveryBySender ดึงข้อมูลรายการจัดส่งตาม sender_id
func GetDeliveryBySender(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ดึง sender_id จาก URL path
		vars := mux.Vars(r)
		senderID, ok := vars["sender_id"]
		if !ok || senderID == "" {
			http.Error(w, "Missing sender ID", http.StatusBadRequest)
			return
		}

		// Query ข้อมูลการจัดส่งพร้อมรายการสินค้าที่เกี่ยวข้องกับ sender_id
		query := `
           SELECT 
                s.shipments, 
                s.sender_id, 
                s.receiver_id, 
                s.rider_id, 
                s.status,
                si.iid,
                si.description,
                si.image
            FROM 
                Shipments s
            JOIN 
                Shipment_Items si 
            ON 
                s.shipments = si.shipment_id
            WHERE 
                s.sender_id = ?
        `

		// Prepare the response structure
		var deliveries []ShipmentDetail

		rows, err := db.Query(query, senderID)
		if err != nil {
			http.Error(w, "Failed to retrieve delivery data", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// ใช้ map เก็บข้อมูลรายการจัดส่งที่มี shipments เดียวกัน
		deliveriesMap := make(map[int]*ShipmentDetail)

		for rows.Next() {
			var shipmentID int
			var delivery ShipmentDetail
			var item ShipmentItem

			// แก้ไขการสแกนให้ตรงตามลำดับ
			err := rows.Scan(&shipmentID, &delivery.SenderID, &delivery.ReceiverID, &delivery.RiderID, &delivery.Status, &item.IID, &item.Description, &item.Image)
			if err != nil {
				log.Printf("Error scanning shipment data: %v", err)
				http.Error(w, "Failed to scan shipment data", http.StatusInternalServerError)
				return
			}

			// เช็คว่ามี delivery ของ shipments นี้แล้วหรือยัง ถ้ามีแล้วให้เพิ่ม items
			if existingDelivery, found := deliveriesMap[shipmentID]; found {
				existingDelivery.Items = append(existingDelivery.Items, item)
			} else {
				delivery.ShipmentID = shipmentID
				delivery.Items = []ShipmentItem{item}
				deliveriesMap[shipmentID] = &delivery
			}
		}

		// เพิ่มข้อมูลจาก map ไปยัง slice ของ deliveries
		for _, delivery := range deliveriesMap {
			deliveries = append(deliveries, *delivery)
		}

		// เช็คว่าพบข้อมูลหรือไม่
		if len(deliveries) == 0 {
			http.Error(w, "No shipments found for this sender", http.StatusNotFound)
			return
		}

		// ส่งข้อมูลกลับในรูปแบบ JSON
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(deliveries); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
