package models

type GroupModel struct {
	Id          string
	Owner       string
	Image       string
	Name        string
	About       string
	CreatedDate string
	NoOfMembers int
	IsBanned    bool
	BanTime     string
	Members     []string
}

type GroupMessageModel struct {
	MsgId    string
	GroupId  string
	SenderId string
	Content  string
	Time     string
}

type GroupMembersModel struct {
	MPhone   string `json:"phone"`
	MName    string `json:"name"`
	MAvatar  string `json:"avatar"`
	MIsAdmin bool   `json:"is_admin"`
}
