package models

type UserModel struct {
	PageTitle     string
	UserID        string
	UserAvatarUrl string
	UserName      string
	UserPhone     string
	UserPass      string
	UserBlocked   string
}

var userVal *UserModel

func InitUserModel(model UserModel) *UserModel {
	userVal = &UserModel{
		PageTitle:     model.PageTitle,
		UserAvatarUrl: model.UserAvatarUrl,
		UserName:      model.UserName,
		UserPhone:     model.UserPhone,
		UserPass:      model.UserPass,
	}
	return userVal
}

func ReturnUserModel() *UserModel {
	return userVal
}
