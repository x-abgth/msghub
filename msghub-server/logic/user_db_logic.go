package logic

import (
	"errors"
	"fmt"
	"log"
	"msghub-server/models"
	"msghub-server/repository"
	"msghub-server/utils"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

type UserDb struct {
	userData repository.User
	groupMsg repository.GroupMessage
	err      error
}

// MigrateUserDb :  Creates table for user according the struct User
func (u *UserDb) MigrateUserDb(db *gorm.DB) error {
	err := db.AutoMigrate(&repository.User{})
	return err
}

func (u *UserDb) UserLoginLogic(phone, password string) (bool, error) {

	var (
		count    int
		userData models.UserModel
	)
	count, userData, u.err = u.userData.GetUserDataUsingPhone(phone)
	if u.err != nil {
		return false, errors.New("you don't have an account, Please register")
	}

	// Check the value is isBlocked and if string convert to bool using if()
	if userData.UserBlocked {
		if userData.BlockDur == "permanent" {
			return false, errors.New("you have been permanently blocked from this website")
		} else {
			t, err := time.Parse("02-01-2006 3:04:05 PM", userData.BlockDur)
			if err != nil {
				log.Println(err)
				return false, errors.New("an unknown error occurred, but you're blocked")
			}

			if float64(t.Sub(time.Now())) < 0.009 {
				err := u.userData.UndoAdminBlockRepo(phone)
				if err != nil {
					return false, errors.New("an unknown error occurred")
				}
				log.Println("Time diff = ", t.Sub(time.Now()))
				return true, nil
			}

			exp := strings.Split(userData.BlockDur, " ")
			return false, errors.New("you have been blocked till " + exp[0] + " from this website")
		}
	} else if count < 1 {
		return false, errors.New("you don't have an account, Please register")
	} else if count > 1 {
		return false, errors.New("something went wrong. Try login again")
		// SHOULD DELETE EXTRA REGISTERED NUMBER!
	} else {
		if utils.CheckPasswordMatch(password, userData.UserPass) {

			models.InitUserModel(userData)
			return true, nil
		} else {
			return false, errors.New("invalid phone number or password")
		}
	}
}

func (u *UserDb) UserDuplicationStatsAndSendOtpLogic(phone string) bool {

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

func (u *UserDb) UserValidateOtpLogic(phone, otp string) bool {
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

func (u *UserDb) UserRegisterLogic(name, phone, pass string) bool {
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

func (u *UserDb) CheckUserRegisterOtpLogic(otp, name, phone, pass string) (bool, string) {
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

func (u *UserDb) GetDataForDashboardLogic(phone string) (models.UserDashboardModel, error) {

	defer func() {
		if e := recover(); e != nil {

			fmt.Println(e)
		}
	}()

	// Got recent personal chats
	personalMessages, err := u.userData.GetRecentChatList(phone)
	if err != nil {
		return models.UserDashboardModel{}, err
	}

	// Get recent group chats
	userGroups, err := u.userData.GetGroupForUser(phone)
	if err != nil {
		return models.UserDashboardModel{}, err
	}

	// assign it to dashboard model
	var recents []models.RecentChatModel

	if len(personalMessages) > 0 {
		// Recent chats for personal messages
		for i := range personalMessages {
			var recentData models.RecentChatModel

			msgSentTime, err := time.Parse("02 Jan 2006 3:04:05 PM", personalMessages[i].Time)
			if err != nil {
				return models.UserDashboardModel{}, err
			}
			diff := time.Now().Sub(msgSentTime)

			if personalMessages[i].From == phone {

				// Get user datas like dp, name
				_, usersData, err := u.userData.GetUserDataUsingPhone(personalMessages[i].To)
				if err != nil {
					return models.UserDashboardModel{}, err
				}

				recentData = models.RecentChatModel{
					Content: models.RecentMessages{
						Id:          personalMessages[i].To,
						Name:        usersData.UserName,
						Avatar:      usersData.UserAvatarUrl,
						LastMsg:     personalMessages[i].Content,
						LastMsgTime: personalMessages[i].Time,
					},
					Sender:  personalMessages[i].From,
					IsGroup: false,
					Order:   float64(diff),
				}
			} else {
				// Get user datas like dp, name
				_, usersData, err := u.userData.GetUserDataUsingPhone(personalMessages[i].From)
				if err != nil {
					return models.UserDashboardModel{}, err
				}

				recentData = models.RecentChatModel{
					Content: models.RecentMessages{
						Id:          personalMessages[i].From,
						Name:        usersData.UserName,
						Avatar:      usersData.UserAvatarUrl,
						LastMsg:     personalMessages[i].Content,
						LastMsgTime: personalMessages[i].Time,
					},
					Sender:  personalMessages[i].From,
					IsGroup: false,
					Order:   float64(diff),
				}
			}
			recents = append(recents, recentData)
		}
	}

	//  GETTING GROUP MESSAGES
	var groupMessages []models.GrpMsgModel
	for i := range userGroups {
		data, err := u.groupMsg.GetRecentGroupMessages(userGroups[i])
		if err != nil {
			return models.UserDashboardModel{}, err
		} else if data.Id == "" {
			return models.UserDashboardModel{}, err
		}

		groupMessages = append(groupMessages, data)
	}

	if len(groupMessages) > 0 {
		for i := range groupMessages {
			groupSentTime, err := time.Parse("02 Jan 2006 3:04:05 PM", groupMessages[i].Time)
			if err != nil {
				return models.UserDashboardModel{}, err
			}
			diff := time.Now().Sub(groupSentTime)
			recentData := models.RecentChatModel{
				Content: models.RecentMessages{
					Id:          groupMessages[i].Id,
					Name:        groupMessages[i].Name,
					Avatar:      groupMessages[i].Avatar,
					LastMsg:     groupMessages[i].Message,
					LastMsgTime: groupMessages[i].Time,
				},
				Sender:  groupMessages[i].Sender,
				IsGroup: true,
				Order:   float64(diff),
			}
			recents = append(recents, recentData)
		}
	}

	// join personal message array and group message array

	// sort the resultant array
	sort.Slice(recents, func(i, j int) bool {
		return recents[i].Order < recents[j].Order
	})

	// get user details
	userDetails, err1 := u.userData.GetUserData(phone)
	if err1 != nil {
		log.Println("getting userDetails")
		return models.UserDashboardModel{}, err1
	}

	returnData := models.UserDashboardModel{
		UserPhone:      phone,
		UserDetails:    userDetails,
		RecentChatList: recents,
		StoryList:      nil,
	}

	return returnData, nil
}

func (u *UserDb) GetAllUsersLogic(ph string) ([]models.UserModel, error) {
	data, err := u.userData.GetAllUsersData(ph)
	if err != nil {
		return []models.UserModel{}, err
	}
	return data, nil
}

func (u *UserDb) GetUserDataLogic(ph string) (models.UserModel, error) {
	data, err := u.userData.GetUserData(ph)
	if err != nil {
		return models.UserModel{}, err
	}
	return data, nil
}

func (u *UserDb) UpdateUserProfileDataLogic(name, about, image, phone string) error {
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
