package routes

import (
	"github.com/gorilla/mux"
	"log"
	"msghub-server/socket"
	"msghub-server/template"
	"net/http"
)

func InitializeRoutes(theMux *mux.Router, server *socket.WsServer) {

	theMux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		socket.ServeWs(server, w, r)
	})

	userRoutes(theMux)
	adminRoutes(theMux)
	theMux.NotFoundHandler = http.HandlerFunc(noPageHandler)
}

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
