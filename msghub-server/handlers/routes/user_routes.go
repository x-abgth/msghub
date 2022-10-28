package routes

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"msghub-server/handlers"
	"msghub-server/handlers/middlewares"
	"msghub-server/socket"
	jwtPkg "msghub-server/utils/jwt"
	"net/http"
)

func userRoutes(theMux *mux.Router, s *socket.WsServer) {
	hub := &socket.Hub{
		Clients:    make(map[string]map[*socket.GClient]bool),
		Register:   make(chan *socket.GClient),
		Unregister: make(chan *socket.GClient),
		Broadcast:  make(chan *socket.WSMessage),
	}
	go hub.Run()

	handlerInfo := handlers.InformationHelper{}

	theMux.HandleFunc("/register", handlerInfo.UserRegisterHandler).Methods("POST")
	theMux.HandleFunc("/register/otp/getotp", handlerInfo.UserOtpPageHandler).Methods("GET")
	theMux.HandleFunc("/register/otp/getotp", handlerInfo.UserVerifyRegisterOtpHandler).Methods("POST")

	// login and register functions
	theMux.HandleFunc("/", middlewares.UserAuthorizationBeforeLogin(handlerInfo.UserLoginHandler)).Methods("GET")
	theMux.HandleFunc("/", handlerInfo.UserLoginCredentialsHandler).Methods("POST")
	theMux.HandleFunc("/login/otp/getphone", handlerInfo.UserLoginWithOtpPhonePageHandler).Methods("GET")
	theMux.HandleFunc("/login/otp/getphone", handlerInfo.UserLoginOtpPhoneValidateHandler).Methods("POST")
	theMux.HandleFunc("/login/otp/getotp", handlerInfo.UserOtpPageHandler).Methods("GET")
	theMux.HandleFunc("/login/otp/getotp", handlerInfo.UserVerifyLoginOtpHandler).Methods("POST")
	theMux.HandleFunc("/user/dashboard", middlewares.UserAuthorizationAfterLogin(handlerInfo.UserDashboardHandler)).Methods("GET")
	theMux.HandleFunc("/user/dashboard/people", middlewares.UserAuthorizationAfterLogin(handlerInfo.UserShowPeopleHandler)).Methods("GET")
	theMux.HandleFunc("/user/dashboard/new-chat-started/{target}", middlewares.UserAuthorizationAfterLogin(handlerInfo.UserNewChatStartedHandler)).Methods("GET")
	theMux.HandleFunc("/user/dashboard/user-profile", handlerInfo.UserProfilePageHandler).Methods("GET")
	theMux.HandleFunc("/user/dashboard/user-profile", handlerInfo.UserProfileUpdateHandler).Methods("POST")
	theMux.HandleFunc("/user/dashboard/create-group", handlerInfo.UserCreateGroup).Methods("POST")
	theMux.HandleFunc("/user/dashboard/add-group-members", handlerInfo.UserAddGroupMembers).Methods("GET")
	theMux.HandleFunc("/user/dashboard/group-created-finally", handlerInfo.UserGroupCreationHandler).Methods("POST")
	theMux.HandleFunc("/user/dashboard/chat-selected", handlerInfo.UserNewChatSelectedHandler).Methods("POST")
	theMux.HandleFunc("/user/dashboard/group-chat-selected", handlerInfo.UserGroupChatSelectedHandler).Methods("POST")

	theMux.HandleFunc("/user/logout", handlerInfo.UserLogoutHandler).Methods("GET")

	// WEBSOCKET CONNECTIONS

	// For personal messaging
	theMux.HandleFunc("/ws/{target}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("--------- IN /WS/TARGET HANDLER FUNCTION ------------")

		defer func() {
			if e := recover(); e != nil {
				log.Println(e)
				http.Redirect(w, r, "/", http.StatusSeeOther)
			}
		}()

		vars := mux.Vars(r)
		target := vars["target"]

		fmt.Println(target)

		c, err1 := r.Cookie("userToken")
		if err1 != nil {
			if err1 == http.ErrNoCookie {
				panic("No cookie found - " + err1.Error())
			}
			panic(err1.Error())
		}

		claim := jwtPkg.GetValueFromJwt(c) // error
		if claim == nil {
			panic("JWT error happened!")
		}

		socket.ServeWs(claim.User.UserPhone, target, s, w, r)
	})

	// For group messaging
	theMux.HandleFunc("/ws/group/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("--------- IN /WS/TARGET HANDLER FUNCTION ------------")

		defer func() {
			if e := recover(); e != nil {
				log.Println(e)
				http.Redirect(w, r, "/", http.StatusSeeOther)
			}
		}()

		vars := mux.Vars(r)
		target := vars["id"]

		socket.ServeGroupWs(hub, target, w, r)
	})
}
