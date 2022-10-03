package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"msghub-server/handlers/routes"
	"msghub-server/repository"
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
	template.Tpl, err = template.Tpl.ParseGlob("../msghub-client/views/user/*.gohtml")
	template.Tpl.New("partials").ParseGlob("../msghub-client/views/base_partials/*.gohtml")
	template.Tpl.New("admin").ParseGlob("../msghub-client/views/admin/*.gohtml")

	if err != nil {
		log.Fatal(err.Error())
	}
	if err1 != nil {
		log.Fatal(err1.Error())
	}
}

// The application starts from here.
func main() {

	flag.Parse()

	repository.ConnectDb()
	defer repository.SqlDb.Close()

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

	routes.InitializeRoutes(newMux)

	server := &http.Server{Addr: ":8080", Handler: newMux}
	fmt.Println("Starting server on port http://localhost:8080")
	go func() {
		server.ListenAndServe()
	}()

	// The channel is only used because the main goroutine will wait
	// for the other goroutine until the value from channel is received.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	// The value received from the channel is not going to use,
	// so we need to provide a variable for that.
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
