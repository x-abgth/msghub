package socket

type WSMessage struct {
	Type    string   `json:"type"`
	Payload GMessage `json:"payload"`
}

type WSImageMessage struct {
	Type    string        `json:"type"`
	Payload GImageMessage `json:"payload"`
}

type GImageMessage struct {
	Body []byte `json:"body"`
	Time string `json:"time"`
	By   string `json:"by"`
	Room string `json:"room"`
}
