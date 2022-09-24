package main

import (
	"context"
	"fmt"
	"log"
	"msghub-server/database"
	"msghub-server/handlers"
	"msghub-server/socket"
	"msghub-server/template"
	utils "msghub-server/utils/jwt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

func init() {
	var err, err1 error

	utils.InitJwtKey()
	template.Tpl, err = template.Tpl.ParseGlob("../msghub-client/views/*.gohtml")
	template.Tpl.New("partials").ParseGlob("../msghub-client/views/partials/*.gohtml")
	template.Tpl.New("admin").ParseGlob("../msghub-client/views/admin/*.gohtml")

	if err != nil {
		log.Fatal(err.Error())
	}
	if err1 != nil {
		log.Fatal(err1.Error())
	}
}

var hub *socket.Hub

// Chat application server side.
func main() {
	database.ConnectDb()
	defer database.SqlDb.Close()

	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	fmt.Println("Server shutdown successfully!")
}

// This function helps to cleanly shutdown the server
func run() error {
	defer func() { // recovers panic
		if e := recover(); e != nil {
			fmt.Println("Recovered from panic")
		}
	}()
	newMux := mux.NewRouter()
	// serving other files like css, and images using only http package
	fileServe := http.FileServer(http.Dir("../msghub-client/assets/"))
	newMux.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", fileServe))

	handleFuncs(newMux)

	server := &http.Server{Addr: ":8080", Handler: newMux}
	fmt.Println("Starting server on port http://localhost:8080")
	go func() {
		server.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	fmt.Println("\nShutting down ... ")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("server failed to shutdown cleanly: %v", err)
	}

	return nil
}

func handleFuncs(theMux *mux.Router) {
	hub = socket.NewHub()
	go hub.Run()

	// SOCKET FUNCTIONS
	theMux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		socket.ServeWs(hub, w, r)
	})

	// OTHER HANDLERS.
	theMux.HandleFunc("/", handlers.UserLoginHandler)
	theMux.HandleFunc("/admin/login-page", handlers.AdminLoginPageHandler)
	theMux.HandleFunc("/admin/authenticate", handlers.AdminAuthenticateHandler)
	// theMux.HandleFunc("/admin/dashboard", handlers.AdminAuthenticateHandler)

	// login and register functions
	theMux.HandleFunc("/register", handlers.UserRegisterHandler)
	theMux.HandleFunc("/register/otp/getotp", handlers.UserOtpPageHandler)
	theMux.HandleFunc("/register/otp/verify", handlers.UserVerifyRegisterOtpHandler)
	theMux.HandleFunc("/login/credentials", handlers.UserLoginCredentialsHandler)
	theMux.HandleFunc("/login/otp/getphone", handlers.UserLoginWithOtpPhonePageHandler)
	theMux.HandleFunc("/login/otp/validatephone", handlers.UserLoginOtpPhoneValidateHandler)
	theMux.HandleFunc("/login/otp/getotp", handlers.UserOtpPageHandler)
	theMux.HandleFunc("/login/otp/verify", handlers.UserVerifyLoginOtpHandler)
	theMux.HandleFunc("/user/dashboard", handlers.UserDashboardHandler)
	theMux.HandleFunc("/user/logout", handlers.UserLogoutHandler).Methods("GET")
}

/*
	TODOS:
		- Input validation while authentication
	 	- While sign up duplicate phone numbers should not be allowed
		- Sign in with otp
		- OTP validation while Sign up
		- forgot password
		- User dashboard ui
		- Chat ui
		- Socket connection and chatting
*/
