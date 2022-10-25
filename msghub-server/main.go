package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"msghub-server/handlers/routes"
	"msghub-server/models"
	"msghub-server/repository"
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
	var err error

	utils.InitJwtKey()
	template.Tpl, err = template.Tpl.ParseGlob("../msghub-client/views/*.gohtml")
	template.Tpl.New("partials").ParseGlob("../msghub-client/views/base_partials/*.gohtml")
	template.Tpl.New("user").ParseGlob("../msghub-client/views/user/*.gohtml")
	template.Tpl.New("user_partials").ParseGlob("../msghub-client/views/user/user_partials/*.gohtml")
	template.Tpl.New("admin").ParseGlob("../msghub-client/views/admin/*.gohtml")
	template.Tpl.New("admin_partials").ParseGlob("../msghub-client/views/admin/admin_partials/*.gohtml")

	if err != nil {
		log.Fatal(err.Error())
	}
}

// The application starts from here.
func main() {
	flag.Parse()

	repository.ConnectDb()
	defer models.SqlDb.Close()

	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	fmt.Println("Server shutdown successfully!")
}

// This function helps to cleanly shut down the server
func run() error {
	newMux := mux.NewRouter()
	// serving other files like css, and assets using only http package
	fileServe := http.FileServer(http.Dir("../msghub-client/assets/"))
	newMux.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", fileServe))

	// creates a new WsServer
	wsServer := socket.NewWebSocketServer()
	go wsServer.Run()
	routes.InitializeRoutes(newMux, wsServer)

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
