package models

type RecentMessages struct {
	Id          string
	Name        string
	Avatar      string
	LastMsg     string
	LastMsgTime string
}

type RecentChatModel struct {
	Content   RecentMessages
	Sender    string
	IsGroup   bool
	Order     float64
	IsBlocked bool
}

type StoryModel struct {
	UserName   string
	UserPhone  string
	UserAvatar string
	StoryImg   string
}

type UserDashboardModel struct {
	UserPhone      string
	UserDetails    UserModel
	RecentChatList []RecentChatModel
	StoryList      []StoryModel
}
