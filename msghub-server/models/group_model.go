package models

type GroupModel struct {
	Id      string
	Owner   string
	Image   string
	Name    string
	About   string
	Members []string
}

type GroupMessageModel struct {
	MsgId    string
	GroupId  string
	SenderId string
	Content  string
	Time     string
}
