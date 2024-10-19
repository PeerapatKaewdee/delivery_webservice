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

	return r
}
