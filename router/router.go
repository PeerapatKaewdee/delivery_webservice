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
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}).Methods("GET")

	// Register rider route
	r.HandleFunc("/api/rider/register", api.RegisterRider(db)).Methods("POST")
	r.HandleFunc("/api/auth/login", api.LoginUserOrRider(db)).Methods("POST")
	r.HandleFunc("/api/user/register", api.RegisterUser(db)).Methods("POST")
	// Route สำหรับการสร้างการจัดส่ง
	r.HandleFunc("/create-delivery", api.CreateDelivery(db)).Methods("POST")
	r.HandleFunc("/search-user", api.SearchReceiverByPhone(db)).Methods("POST")
	r.HandleFunc("/get/list_user_send/{sender_id}", api.GetDeliveryBySender(db)).Methods("POST")
	r.HandleFunc("/get/rider/{rider_id}", api.GetRider(db)).Methods("POST")

	return r
}
