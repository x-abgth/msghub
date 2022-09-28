package socket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

/*
	HEARTBEAT MECHANISM (PING AND PONG)

	In order to ensure that the TCP channel connection between the client and the server is not disconnected,
	WebSocket uses the heartbeat mechanism to judge the connection status.
	If no response is received within the timeout period,the connection is considered disconnected,
	the connection is closed and resources are released.
*/

// Sender -> reciever : ping
// Receiver	-> sender : pong

const (
	// Max wait time when writing message to peer
	writeWait = 10 * time.Second

	// Max time till next pong from peer
	pongWait = 60 * time.Second

	// Send ping interval, must be less then pong wait time
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 10000
)

var (
	newline = []byte{'\n'}
	// space    = []byte{' '}
	upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
	}
)

// This represents the websocket client at the server
type Client struct {
	conn *websocket.Conn
	// We want to the keep a reference to the WsServer for each client
	wsServer *WsServer
	send     chan []byte
	// To keep track of the rooms this client joins
	rooms map[*Room]bool
	Name  string `json:"name"`
}

func newClient(conn *websocket.Conn, wsServer *WsServer, name string) *Client {
	return &Client{
		conn:     conn,
		wsServer: wsServer,
		Name:     name,
		send:     make(chan []byte),
		rooms:    make(map[*Room]bool),
	}
}

func (client *Client) GetName() string {
	return client.Name
}

// ServeWs handles websocket requests from clients requests.
func ServeWs(wsServer *WsServer, w http.ResponseWriter, r *http.Request) {
	name, ok := r.URL.Query()["name"]

	if !ok || len(name[0]) < 1 {
		log.Println("Url Param 'name' is missing")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// whenever the fuction ServeWs is called a new client is created.
	client := newClient(conn, wsServer, name[0])

	go client.writePump()
	go client.readPump()

	// value of client is updated to the channel register.
	wsServer.register <- client

	fmt.Println("New Client joined the hub!")
	fmt.Println(client)
}

func (client *Client) readPump() {
	defer func() {
		client.wsServer.unregister <- client
		client.conn.Close()
	}()

	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error { client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// start endless read loop, waiting for messages from client.
	for {
		_, jsonMessage, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close error: %v", err.Error())
			}
			break
		}
		client.handleNewMessage(jsonMessage)
		fmt.Println("Reading -- ", string(jsonMessage))
	}
}

/*
	The writePump is also responsible for keeping the connection alive
	by sending ping messages to the client with the interval given in pingPeriod.
	If the client does not respond with a pong, the connection is closed.
*/

func (client *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(client.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-client.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (client *Client) handleNewMessage(jsonMessage []byte) {

	var message Message
	if err := json.Unmarshal(jsonMessage, &message); err != nil {
		log.Printf("Error on unmarshal JSON message %s", err)
	}

	message.Sender = client

	switch message.Action {
	case SendMessageAction:
		// This will send messages to a specific room
		roomName := message.Target.GetName()
		if room := client.wsServer.findRoomByName(roomName); room != nil {
			room.broadcast <- &message
		}
	case JoinRoomAction:
		client.handleJoinRoomMessage(message)
	case LeaveRoomAction:
		client.handleLeaveRoomMessage(message)
	}
}

func (client *Client) handleJoinRoomMessage(message Message) {
	roomName := message.Message

	room := client.wsServer.findRoomByName(roomName)
	if room == nil {
		room = client.wsServer.createRoom(roomName)
	}

	client.rooms[room] = true
	room.register <- client
}

func (client *Client) handleLeaveRoomMessage(message Message) {
	room := client.wsServer.findRoomByName(message.Message)
	if _, ok := client.rooms[room]; ok {
		delete(client.rooms, room)
	}

	room.unregister <- client
}

func (client *Client) disconnect() {
	client.wsServer.unregister <- client
	for room := range client.rooms {
		room.unregister <- client
	}
}
