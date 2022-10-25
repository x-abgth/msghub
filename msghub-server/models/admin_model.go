package models

type AdminModel struct {
	AdminId   string
	AdminName string
}

type AdminDashboardModel struct {
	AdminName      string
	UsersTbContent []UserModel
	AdminTbContent []AdminModel
	GroupTbContent []GroupModel
}
