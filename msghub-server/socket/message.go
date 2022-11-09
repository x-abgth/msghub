package socket

import (
	"encoding/json"
	"fmt"
	"log"
)

const SendMessageAction = "send-message"
const JoinRoomAction = "join-room"
const LeaveRoomAction = "leave-room"
const UserJoinedAction = "user-join"
const UserLeftAction = "user-left"
const JoinRoomPrivateAction = "join-room-private"
const RoomJoinedAction = "room-joined"

type Message struct {
	Action    string  `json:"action"`
	Message   string  `json:"message"`
	Time      string  `json:"time"`
	Target    *Room   `json:"target"`
	Sender    *Client `json:"sender"`
	IsPrivate bool    `json:"is_bool"`
}

type GMessage struct {
	File string `json:"file"`
	Body string `json:"body"`
	Time string `json:"time"`
	By   string `json:"by"`
	Room string `json:"room"`
}

func (message *Message) encode() []byte {
	jsonStr, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}

	return jsonStr
}

func (message *Message) decode(jsonStr []byte) Message {
	isValid := json.Valid(jsonStr)

	if isValid {
		json.Unmarshal(jsonStr, &message)
		return *message
	} else {
		fmt.Println("Json is not valid!")
		return *message
	}
}
