package routes

import (
	"github.com/gorilla/mux"
	"msghub-server/handlers"
	"msghub-server/handlers/middlewares"
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

	theMux.HandleFunc("/user/logout", handlerInfo.UserLogoutHandler).Methods("GET")
}
