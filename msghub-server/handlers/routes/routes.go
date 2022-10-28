package routes

import (
	"log"
	"msghub-server/socket"
	"msghub-server/template"
	"net/http"

	"github.com/gorilla/mux"
)

func InitializeRoutes(theMux *mux.Router, server *socket.WsServer) {
	userRoutes(theMux, server)
	adminRoutes(theMux)
	theMux.NotFoundHandler = http.HandlerFunc(noPageHandler)
}

// 404 Error page handler function
func noPageHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title string
	}{
		Title: "404 Error Page",
	}

	err := template.Tpl.ExecuteTemplate(w, "error_page.gohtml", data)
	if err != nil {
		log.Fatal("Couldn't render the error page handler!")
	}
}
