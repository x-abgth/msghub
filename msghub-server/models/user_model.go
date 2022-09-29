package models

type UserModel struct {
	PageTitle     string
	UserAvatarUrl string
	UserAbout     string
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
		UserAbout:     model.UserAbout,
		UserName:      model.UserName,
		UserPhone:     model.UserPhone,
		UserPass:      model.UserPass,
	}
	return userVal
}

func ReturnUserModel() *UserModel {
	return userVal
}
