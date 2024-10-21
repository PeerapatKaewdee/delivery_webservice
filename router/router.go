package router

import (
	"database/sql"
	"delivery_webservice/api" // Import the api package
	"net/http"

	"github.com/gorilla/mux"
)

func InitRoutes(db *sql.DB) *mux.Router {
	r := mux.NewRouter()

	// Example route
	r.HandleFunc("/api/example", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}).Methods("GET")

	// Register rider route
	r.HandleFunc("/api/rider/register", api.RegisterRider(db)).Methods("POST")
	r.HandleFunc("/api/auth/login", api.LoginUserOrRider(db)).Methods("POST")
	r.HandleFunc("/api/user/register", api.RegisterUser(db)).Methods("POST")
	// Route สำหรับการสร้างการจัดส่ง
	r.HandleFunc("/shipments", api.CreateShipment(db)).Methods("POST")       // สร้างการจัดส่งใหม่
    r.HandleFunc("/shipments", api.UpdateShipment(db)).Methods("PUT")        // อัปเดตการจัดส่ง
    r.HandleFunc("/shipments", api.GetShipment(db)).Methods("GET")  
// Route สำหรับการสร้าง shipment ใหม่


	return r
}
