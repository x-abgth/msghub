package routes

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"msghub-server/socket"
	"msghub-server/template"
	jwtPkg "msghub-server/utils/jwt"
	"net/http"
)

func InitializeRoutes(theMux *mux.Router, server *socket.WsServer) {

	theMux.HandleFunc("/ws/{target}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("--------- IN /WS/TARGET HANDLER FUNCTION ------------")

		vars := mux.Vars(r)
		target := vars["target"]

		fmt.Println(target)

		c, err1 := r.Cookie("userToken")
		if err1 != nil {
			if err1 == http.ErrNoCookie {
				http.Redirect(w, r, "/", http.StatusSeeOther)
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}

		claim := jwtPkg.GetValueFromJwt(c)

		socket.ServeWs(claim.User.UserPhone, target, server, w, r)
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
