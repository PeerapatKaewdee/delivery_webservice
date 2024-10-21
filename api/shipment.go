package api

import (
    "database/sql"
    "encoding/json"
    "net/http"
)

// Shipment represents the shipment structure
type Shipment struct {
    ID             int    `json:"id"`
    OrderID        int    `json:"order_id"`
    TrackingNumber string `json:"tracking_number"`
    Status         string `json:"status"` // e.g. "pending", "shipped", "delivered"
    Carrier        string `json:"carrier"` // e.g. "UPS", "FedEx"
}

// CreateShipmentRequest represents the request structure for creating a shipment
type CreateShipmentRequest struct {
    OrderID        int    `json:"order_id"`
    TrackingNumber string `json:"tracking_number"`
    Status         string `json:"status"`
    Carrier        string `json:"carrier"`
}

// CreateShipment handles the creation of a new shipment
func CreateShipment(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req CreateShipmentRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid input", http.StatusBadRequest)
            return
        }

        // Insert shipment into the database
        query := "INSERT INTO Shipments (order_id, tracking_number, status, carrier) VALUES (?, ?, ?, ?)"
        result, err := db.Exec(query, req.OrderID, req.TrackingNumber, req.Status, req.Carrier)
        if err != nil {
            http.Error(w, "Failed to create shipment", http.StatusInternalServerError)
            return
        }

        // Get the ID of the newly created shipment
        shipmentID, err := result.LastInsertId()
        if err != nil {
            http.Error(w, "Failed to retrieve shipment ID", http.StatusInternalServerError)
            return
        }

        // Respond with the created shipment information
        response := Shipment{
            ID:             int(shipmentID),
            OrderID:        req.OrderID,
            TrackingNumber: req.TrackingNumber,
            Status:         req.Status,
            Carrier:        req.Carrier,
        }

        w.WriteHeader(http.StatusCreated) // Set the HTTP status code to 201 Created
        json.NewEncoder(w).Encode(response)
    }
}

// GetShipment retrieves a shipment by its ID
func GetShipment(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        id := r.URL.Query().Get("id")
        if id == "" {
            http.Error(w, "Missing shipment ID", http.StatusBadRequest)
            return
        }

        var shipment Shipment
        query := "SELECT id, order_id, tracking_number, status, carrier FROM Shipments WHERE id = ?"
        err := db.QueryRow(query, id).Scan(&shipment.ID, &shipment.OrderID, &shipment.TrackingNumber, &shipment.Status, &shipment.Carrier)
        if err != nil {
            if err == sql.ErrNoRows {
                http.Error(w, "Shipment not found", http.StatusNotFound)
            } else {
                http.Error(w, "Failed to retrieve shipment", http.StatusInternalServerError)
            }
            return
        }

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(shipment)
    }
}

// UpdateShipment updates the details of a shipment
func UpdateShipment(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var shipment Shipment
        if err := json.NewDecoder(r.Body).Decode(&shipment); err != nil {
            http.Error(w, "Invalid input", http.StatusBadRequest)
            return
        }

        // Update the shipment in the database
        query := "UPDATE Shipments SET tracking_number = ?, status = ?, carrier = ? WHERE id = ?"
        _, err := db.Exec(query, shipment.TrackingNumber, shipment.Status, shipment.Carrier, shipment.ID)
        if err != nil {
            http.Error(w, "Failed to update shipment", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(shipment)
    }
}
