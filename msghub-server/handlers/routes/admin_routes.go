package routes

import (
	"github.com/gorilla/mux"
	"msghub-server/handlers"
)

func adminRoutes(theMux *mux.Router) {
	// OTHER HANDLERS.
	theMux.HandleFunc("/admin/login-page", handlers.AdminLoginPageHandler).Methods("GET")
	theMux.HandleFunc("/admin/login-page", handlers.AdminAuthenticateHandler).Methods("POST")
	// theMux.HandleFunc("/admin/dashboard", handlers.AdminAuthenticateHandler)
}
