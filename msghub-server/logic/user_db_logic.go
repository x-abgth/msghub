package logic

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"msghub-server/models"
	"msghub-server/repository"
	"msghub-server/utils"
)

type UserDb struct {
	userData repository.User
	err      error
}

// MigrateUserDb :  Creates table for user according the struct User
func (u UserDb) MigrateUserDb(db *gorm.DB) error {
	err := db.AutoMigrate(&repository.User{})
	return err
}

func (u UserDb) UserLoginLogic(phone, password string) (bool, error) {

	var count int
	count, u.userData, u.err = u.userData.GetUserDataUsingPhone(phone)
	if u.err != nil {
		return false, errors.New("you don't have an account, Please register")
	}

	// Check the value is isBlocked and if string convert to bool using if()
	if u.userData.IsBlocked {
		return false, errors.New("you are temporarily blocked from this application")
	} else if count < 1 {
		return false, errors.New("you don't have an account, Please register")
	} else if count > 1 {
		return false, errors.New("something went wrong. Try login again")
		// SHOULD DELETE EXTRA REGISTERED NUMBER!
	} else {
		if utils.CheckPasswordMatch(password, u.userData.UserPassword) {
			var user models.UserModel

			var blank = ""
			if u.userData.UserAvatar == nil {
				u.userData.UserAvatar = &blank
			}
			user = models.UserModel{
				UserAvatarUrl: *u.userData.UserAvatar,
				UserAbout:     u.userData.UserAbout,
				UserName:      u.userData.UserName,
				UserPhone:     phone,
			}

			models.InitUserModel(user)
			return true, nil
		} else {
			return false, errors.New("invalid phone number or password")
		}
	}
}

func (u UserDb) UserDuplicationStatsAndSendOtpLogic(phone string) bool {

	var count int
	count, _, u.err = u.userData.GetUserDataUsingPhone(phone)
	if u.err != nil {
		return false
	}

	if count == 1 {
		status := utils.SendOtp(phone)
		if status {
			data := models.IncorrectOtpModel{
				PhoneNumber: phone,
				IsLogin:     true,
			}
			models.InitOtpErrorModel(data)
			return true
		} else {
			errorStr := models.IncorrectPhoneModel{
				ErrorStr: "Couldn't send OTP to this number!",
			}
			models.InitPhoneErrorModel(errorStr)
			return false
		}
	} else {
		errorStr := models.IncorrectPhoneModel{
			ErrorStr: "The phone number you entered is not registered!",
		}
		models.InitPhoneErrorModel(errorStr)
		return false
	}
}

func (u UserDb) UserValidateOtpLogic(phone, otp string) bool {
	status := utils.CheckOtp(phone, otp)
	if status {
		return true
	} else {
		data := models.IncorrectOtpModel{
			ErrorStr:    "The OTP is incorrect. Try again",
			PhoneNumber: phone,
			IsLogin:     true,
		}
		models.InitOtpErrorModel(data)
		return false
	}
}

func (u UserDb) UserRegisterLogic(name, phone, pass string) bool {
	total := u.userData.UserDuplicationStatus(phone)
	encryptedFormPassword, err := utils.HashEncrypt(pass)

	if err != nil {
		alm := models.AuthErrorModel{
			ErrorStr: "The password is too weak. Please enter a strong password.",
		}
		models.InitAuthErrorModel(alm)
		return false
	} else {
		if total > 0 {
			alm := models.AuthErrorModel{
				ErrorStr: "Account already exist with this number. Try Login method.",
			}
			models.InitAuthErrorModel(alm)
			return false
		} else {
			status := utils.SendOtp(phone)
			if status {
				data := models.IncorrectOtpModel{
					PhoneNumber: phone,
					IsLogin:     false,
				}
				user := models.UserModel{
					UserName:  name,
					UserPhone: phone,
					UserPass:  encryptedFormPassword,
				}
				models.InitUserModel(user)
				models.InitOtpErrorModel(data)
				return true
			} else {
				alm := models.AuthErrorModel{
					ErrorStr: "Couldn't send OTP to this number. Please check the number.",
				}
				models.InitAuthErrorModel(alm)
				return false
			}
		}
	}
}

func (u UserDb) CheckUserRegisterOtpLogic(otp, name, phone, pass string) (bool, string) {
	status := utils.CheckOtp(phone, otp)
	if status {
		done, alert := u.userData.RegisterUser(name, phone, pass)
		if done {
			return true, ""
		}
		alm := models.AuthErrorModel{
			ErrorStr: alert.Error(),
		}
		models.InitAuthErrorModel(alm)
		return false, "login"
	} else {
		data := models.IncorrectOtpModel{
			ErrorStr:    "The OTP is incorrect. Try again",
			PhoneNumber: phone,
			IsLogin:     false,
		}
		models.InitOtpErrorModel(data)
		return false, "otp"
	}
}

func (u UserDb) GetDataForDashboardLogic(phone string) (models.UserDashboardModel, error) {
	userDetails, err1 := u.userData.GetUserData(phone)
	if err1 != nil {
		return models.UserDashboardModel{}, err1
	}

	recentChatList, err2 := u.userData.GetRecentChatList(phone)
	if err2 != nil {
		return models.UserDashboardModel{}, err2
	}

	data := models.UserDashboardModel{
		UserPhone:      phone,
		UserDetails:    userDetails,
		RecentChatList: recentChatList,
		StoryList:      nil,
	}

	return data, nil
}

func (u UserDb) GetAllUsersLogic(ph string) ([]models.UserModel, error) {
	data, err := u.userData.GetAllUsersData(ph)
	if err != nil {
		return []models.UserModel{}, err
	}
	return data, nil
}

func (u UserDb) GetUserDataLogic(ph string) (models.UserModel, error) {
	data, err := u.userData.GetUserData(ph)
	if err != nil {
		return models.UserModel{}, err
	}
	return data, nil
}

func (u UserDb) UpdateUserProfileDataLogic(name, about, image, phone string) error {
	data := models.UserModel{
		UserName:      name,
		UserAbout:     about,
		UserAvatarUrl: image,
		UserPhone:     phone,
	}

	fmt.Println(data)

	err := u.userData.UpdateUserData(data)
	if err != nil {
		return err
	}
	return nil
}
