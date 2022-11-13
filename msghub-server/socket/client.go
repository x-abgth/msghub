package socket

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"log"
	"msghub-server/models"
	"msghub-server/utils"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/gorilla/websocket"
)

/*
	HEARTBEAT MECHANISM (PING AND PONG)

	In order to ensure that the TCP channel connection between the client and the server is not disconnected,
	WebSocket uses the heartbeat mechanism to judge the connection status.
	If no response is received within the timeout period,the connection is considered disconnected,
	the connection is closed and resources are released.
*/

// Sender -> receiver : ping
// Receiver	-> sender : pong

const (
	// Max wait time when writing message to peer
	writeWait = 10 * time.Second

	// Max time till next pong from peer
	pongWait = 60 * time.Second

	// Send ping interval, must be less then pong wait time
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 10485760
)

var (
	newline  = []byte{'\n'}
	upgrader = websocket.Upgrader{
		ReadBufferSize:  5242880,
		WriteBufferSize: 5242880,
	}
)

// Client - This represents the websocket client at the server
type Client struct {
	ID       string `json:"id"`
	conn     *websocket.Conn
	IsJoined bool
	// We want to the keep a reference to the WsServer for each client
	wsServer *WsServer
	send     chan *WSMessage
	Hub      *Hub
	// To keep track of the rooms this client joins
	Name string `json:"name"`
}

type GClient struct {
	// websocket connection
	Conn *websocket.Conn

	// Weather user is joined in group or not
	IsJoined bool

	// send channel
	Send chan *WSMessage

	// Hub
	Hub *Hub

	// Room name
	Room string
}

func newClient(conn *websocket.Conn, wsServer *WsServer, phone string) *Client {
	return &Client{
		ID:       phone,
		conn:     conn,
		wsServer: wsServer,
		send:     make(chan *WSMessage, 256),
	}
}

func (client *Client) GetName() string {
	return client.Name
}

// ServeWs handles websocket requests from clients requests.
// The ServeWs function will be called from routes.go
func ServeWs(phone, target string, wsServer *WsServer, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// whenever the function ServeWs is called a new client is created.
	client := newClient(conn, wsServer, phone)

	models.ClientID = client.ID
	models.TargetID = target

	go client.writePump()
	go client.readPump()

	// value of client is updated to the channel register.
	wsServer.register <- client

	fmt.Println("New Client joined the hub!")
	fmt.Println(client)
}

// ------------------------------------------------------------------------------
// -------------------- PRIVATE CHAT READ AND WRITE PUMP ------------------------
// ------------------------------------------------------------------------------

func (client *Client) readPump() {
	fmt.Println("At read pump")
	defer func() {
		client.wsServer.unregister <- client
		client.conn.Close()
	}()

	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error { client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// start endless read loop, waiting for messages from client.
	for {

		var message WSMessage

		err := client.conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close error: %v", err)
			}
			break
		}

		if message.Payload.Room == "admin" {
			break
		}

		switch message.Type {
		case "join":
			client.IsJoined = true
			var m *WSMessage = &WSMessage{
				Type: message.Type,
				Payload: GMessage{
					By:   message.Payload.By,
					Room: message.Payload.Room,
				},
			}
			log.Println("The client has joined")
			client.wsServer.broadcast <- m

		case "left":
			client.IsJoined = false
			var m *WSMessage = &WSMessage{
				Type: message.Type,
				Payload: GMessage{
					By:   message.Payload.By,
					Room: message.Payload.Room,
				},
			}
			log.Println("The client has left")
			client.wsServer.broadcast <- m

		case "message":
			var m *WSMessage = &WSMessage{
				Type: "message",
				Payload: GMessage{
					Body: message.Payload.Body,
					Time: message.Payload.Time,
					By:   message.Payload.By,
					Room: message.Payload.Room,
				},
			}
			fmt.Println("Read Message ---------------------- ", m.Payload.Room)
			client.wsServer.broadcast <- m

		case "image":
			idx := strings.Index(message.Payload.Body, ";base64,")
			if idx < 0 {
				panic("Error1")
			}

			b64data := message.Payload.Body[strings.IndexByte(message.Payload.Body, ',')+1:]

			// This is not working as expected.
			unbased, err := base64.StdEncoding.DecodeString(string(b64data))
			if err != nil {
				log.Println("ERROR IS HAPPENING IN DECODE STRING")
				log.Println(err)
			}
			res := bytes.NewReader(unbased)
			path, _ := os.Getwd()

			newPath := filepath.Join(path + "/storage")
			err = os.MkdirAll(newPath, os.ModePerm)
			if err != nil {
				log.Println(err)
			}
			uid := uuid.New()

			pngI, _, errPng := image.Decode(res)
			var fileUrl string
			if errPng == nil {
				f, _ := os.OpenFile(newPath+"/"+uid.String()+".png", os.O_WRONLY|os.O_CREATE, 0777)
				png.Encode(f, pngI)

				file, err := os.Open(newPath + "/" + uid.String() + ".png")
				if err != nil {
					panic(err)
				}

				fileUrl = utils.StoreThisFileInBucket("pm_chat_file/", uid.String(), file)

				file.Close()
				os.Remove(newPath + "/" + uid.String() + ".png")
			} else {
				fmt.Println("Error png is having error")
				fmt.Println(errPng.Error())
			}

			var m *WSMessage = &WSMessage{
				Type: "image",
				Payload: GMessage{
					Body: fileUrl,
					Time: message.Payload.Time,
					By:   message.Payload.By,
					Room: message.Payload.Room,
				},
			}
			client.wsServer.broadcast <- m

		case "typing":
			var m *WSMessage = &WSMessage{
				Type: message.Type,
				Payload: GMessage{
					By:   message.Payload.By,
					Room: message.Payload.Room,
				},
			}
			log.Println("Typing:", m)
			client.wsServer.broadcast <- m

		case "stoptyping":
			var m *WSMessage = &WSMessage{
				Type: message.Type,
				Payload: GMessage{
					By:   message.Payload.By,
					Room: message.Payload.Room,
				},
			}
			log.Println("StopTyping:", m)
			client.wsServer.broadcast <- m
		}
	}
}

/*
	The writePump is also responsible for keeping the connection alive
	by sending ping messages to the client with the interval given in pingPeriod.
	If the client does not respond with a pong, the connection is closed.
*/

func (client *Client) writePump() {
	fmt.Println("At write Pump")

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
				log.Println("Not okay")
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := client.conn.WriteJSON(message)
			if err != nil {
				log.Println("Error while writing message:", err)
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

/// Group chat

func ServeGroupWs(hub *Hub, room string, w http.ResponseWriter, r *http.Request) {
	gConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading request to websocket connection: ", err)
	}

	client := &GClient{
		Conn: gConn,
		Send: make(chan *WSMessage),
		Hub:  hub,
		Room: room,
	}

	client.Hub.Register <- client

	go client.GroupReadPump()
	go client.GroupWritePump()
}

func (c *GClient) GroupReadPump() {
	defer func() {
		// unregister client
		c.Hub.Unregister <- c
		// close channel
		close(c.Send)
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		var message WSMessage

		err := c.Conn.ReadJSON(&message)
		if err != nil {
			fmt.Println("Error while reading websocket message: ", err)
			return
		}
		switch message.Type {
		case "join":
			c.IsJoined = true
			var m *WSMessage = &WSMessage{
				Type: message.Type,
				Payload: GMessage{
					By:   message.Payload.By,
					Room: message.Payload.Room,
				},
			}
			log.Println("The client has joined")
			c.Hub.Broadcast <- m

		case "left":
			c.IsJoined = false
			var m *WSMessage = &WSMessage{
				Type: message.Type,
				Payload: GMessage{
					By:   message.Payload.By,
					Room: message.Payload.Room,
				},
			}
			log.Println("The client has left")
			c.Hub.Broadcast <- m

		case "message":
			var m *WSMessage = &WSMessage{
				Type: "message",
				Payload: GMessage{
					Body: message.Payload.Body,
					Time: message.Payload.Time,
					By:   message.Payload.By,
					Room: message.Payload.Room,
				},
			}
			fmt.Println("Read Message ---------------------- ", m.Payload.Room)
			c.Hub.Broadcast <- m

		case "image":
			idx := strings.Index(message.Payload.Body, ";base64,")
			if idx < 0 {
				panic("Error1")
			}

			b64data := message.Payload.Body[strings.IndexByte(message.Payload.Body, ',')+1:]

			// This is not working as expected.
			unbased, err := base64.StdEncoding.DecodeString(string(b64data))
			if err != nil {
				log.Println("ERROR IS HAPPENING IN DECODE STRING")
				log.Println(err)
			}
			res := bytes.NewReader(unbased)
			path, _ := os.Getwd()

			newPath := filepath.Join(path + "/storage")
			err = os.MkdirAll(newPath, os.ModePerm)
			if err != nil {
				log.Println(err)
			}
			uid := uuid.New()

			pngI, _, errPng := image.Decode(res)
			var fileUrl string
			if errPng == nil {
				f, _ := os.OpenFile(newPath+"/"+uid.String()+".png", os.O_WRONLY|os.O_CREATE, 0777)
				png.Encode(f, pngI)

				file, err := os.Open(newPath + "/" + uid.String() + ".png")
				if err != nil {
					panic(err)
				}

				fileUrl = utils.StoreThisFileInBucket("group_chat_file/", uid.String(), file)

				file.Close()
				os.Remove(newPath + "/" + uid.String() + ".png")
			} else {
				fmt.Println("Error png is having error")
				fmt.Println(errPng.Error())
			}

			var m *WSMessage = &WSMessage{
				Type: "image",
				Payload: GMessage{
					Body: fileUrl,
					Time: message.Payload.Time,
					By:   message.Payload.By,
					Room: message.Payload.Room,
				},
			}
			c.Hub.Broadcast <- m

		case "typing":
			var m *WSMessage = &WSMessage{
				Type: message.Type,
				Payload: GMessage{
					By:   message.Payload.By,
					Room: message.Payload.Room,
				},
			}
			log.Println("Typing:", m)
			c.Hub.Broadcast <- m

		case "stoptyping":
			var m *WSMessage = &WSMessage{
				Type: message.Type,
				Payload: GMessage{
					By:   message.Payload.By,
					Room: message.Payload.Room,
				},
			}
			log.Println("StopTyping:", m)
			c.Hub.Broadcast <- m
		}
	}
}

func (c *GClient) GroupWritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	fmt.Println("Group writing section")
	for {
		select {
		case m, ok := <-c.Send:
			if !ok {
				log.Println("Send Channel was closed")
				return
			}

			err := c.Conn.WriteJSON(m)
			if err != nil {
				log.Println("Error while writing message:", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
