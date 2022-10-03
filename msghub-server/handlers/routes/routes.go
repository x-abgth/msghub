package routes

import (
	"github.com/gorilla/mux"
	"msghub-server/socket"
	"net/http"
)

func InitializeRoutes(theMux *mux.Router) {
	// creates a new WsServer
	wsServer := socket.NewWebSocketServer()
	go wsServer.Run()

	theMux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		socket.ServeWs(wsServer, w, r)
	})

	userRoutes(theMux)
	adminRoutes(theMux)
}
