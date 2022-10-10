package models

type RecentChatModel struct {
	UserName    string
	UserPhone   string
	UserAvatar  string
	LastMsg     string
	LastMsgTime string
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
