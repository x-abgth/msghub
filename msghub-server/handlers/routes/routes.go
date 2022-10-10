package routes

import (
	"github.com/gorilla/mux"
	"log"
	"msghub-server/socket"
	"msghub-server/template"
	jwtPkg "msghub-server/utils/jwt"
	"net/http"
)

func InitializeRoutes(theMux *mux.Router, server *socket.WsServer) {

	theMux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		c, err1 := r.Cookie("userToken")
		if err1 != nil {
			if err1 == http.ErrNoCookie {
				http.Redirect(w, r, "/", http.StatusSeeOther)
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}

		claim := jwtPkg.GetValueFromJwt(c)

		socket.ServeWs(claim.User.UserPhone, server, w, r)
	})

	userRoutes(theMux)
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
