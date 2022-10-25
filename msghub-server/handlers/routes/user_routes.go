package routes

import (
	"msghub-server/handlers"
	"msghub-server/handlers/middlewares"

	"github.com/gorilla/mux"
)

func userRoutes(theMux *mux.Router) {
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
}
