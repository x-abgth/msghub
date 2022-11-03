package socket

import (
	"encoding/json"
	"fmt"
	"log"
	"msghub-server/logic"
	"msghub-server/models"
	"os"
)

//	Because our ChatServer acts like a hub for connecting the parts of our chat application,
//	we will use it to keep track of all the rooms that will be created.

type WsServer struct {
	users      []models.UserModel
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	rooms      map[*Room]bool
}

// NewWebSocketServer :- First we create this server.
func NewWebSocketServer() *WsServer {
	return &WsServer{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		rooms:      make(map[*Room]bool),
	}
}

// Run our websocket server, accepting various requests
// This function will run finely and listens to the channels
func (server *WsServer) Run() {
	for {
		select {
		case client := <-server.register:
			server.registerClient(client)
		case client := <-server.unregister:
			server.unregisterClient(client)
		case message := <-server.broadcast:
			server.broadcastToClients(message)
		}
	}
}

// If a client is joined we will make the map value to true.
func (server *WsServer) registerClient(client *Client) {
	//room := NewRoom(client.Name, client.ID, true)
	//room.RunRoom()
	//server.listOnlineClients(client)
	server.clients[client] = true
}

// If the client is left from the socket, we will delete the client key and value.
func (server *WsServer) unregisterClient(client *Client) {
	fmt.Println("unregistered")
	//server.notifyClientLeft(client)
	delete(server.clients, client)
}

// If the client send a message, it broadcasts to all the other users
func (server *WsServer) broadcastToClients(message []byte) {
	defer func() {
		if e := recover(); e != nil {
			log.Println("Error happened in sending message")
			log.Println(e)
			os.Exit(1)
		}
	}()

	var msgModel logic.MessageDb

	user := server.findClientByID(models.ClientID)
	target := server.findClientByID(models.TargetID)

	if user != nil {
		fmt.Println("-----------MESSAGE ---------------")
		fmt.Println(message)
		var model models.MessageModel

		err := json.Unmarshal(message, &model)
		if err != nil {
			panic(err.Error())
		}

		model.Status = logic.IS_SENT

		msgModel.StorePersonalMessagesLogic(model)

		user.send <- message
		if target != nil {
			target.send <- message
		}
	}
}

func (server *WsServer) notifyClientJoined(client *Client) {
	message := &Message{
		Action: UserJoinedAction,
		Sender: client,
	}

	server.broadcastToClients(message.encode())
}

func (server *WsServer) notifyClientLeft(client *Client) {
	message := &Message{
		Action: UserLeftAction,
		Sender: client,
	}

	server.broadcastToClients(message.encode())
}

func (server *WsServer) listOnlineClients(client *Client) {
	fmt.Println("Inside listonlineclients -- ")
	for existingClient := range server.clients {
		message := &Message{
			Action: UserJoinedAction,
			Sender: existingClient,
		}
		fmt.Println(message)
		client.send <- message.encode()
	}
}

func (server *WsServer) findRoomByName(name string) *Room {
	var foundRoom *Room
	for room := range server.rooms {
		if room.GetName() == name {
			foundRoom = room
			break
		}
	}
	return foundRoom
}

func (server *WsServer) createRoom(name string, private bool) *Room {
	room := NewRoom(name, "", private)
	go room.RunRoom()
	server.rooms[room] = true

	return room
}

func (server *WsServer) findRoomByID(ID string) *Room {
	var foundRoom *Room
	for room := range server.rooms {
		if room.GetId() == ID {
			foundRoom = room
			break
		}
	}
	return foundRoom
}

func (server *WsServer) findClientByID(ID string) *Client {
	var foundClient *Client
	for client := range server.clients {
		if client.ID == ID {
			foundClient = client
			break
		}
	}
	return foundClient
}
