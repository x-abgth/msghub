package models

type MessageModel struct {
	Content string `json:"message"`
	From    string `json:"from"`
	Time    string `json:"time"`
	Order   float64
}
