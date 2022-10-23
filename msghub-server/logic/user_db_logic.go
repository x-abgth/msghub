package logic

import (
	"errors"
	"fmt"
	"log"
	"msghub-server/models"
	"msghub-server/repository"
	"msghub-server/utils"
	"sort"
	"strconv"
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

	userDetails, err1 := u.userData.GetUserData(phone)
	if err1 != nil {
		log.Println("getting userDetails")
		return models.UserDashboardModel{}, err1
	}

	recentChatList, err2 := u.userData.GetRecentChatList(phone)
	if err2 != nil {
		log.Println("Getting recentChatList")
		return models.UserDashboardModel{}, err2
	}
	fmt.Println(recentChatList)

	/*
		Also should add group recent chats.
		Therefore, first get all the groups that this user is in
		then find get all the message of that each group,
		remove all the duplicates according to the time
	*/
	// TODO: Get all the groups the user is in (as an array)
	groupsArr, err2 := u.userData.GetGroupForUser(phone)
	if err2 != nil {
		log.Println("Getting groupsArr")
		return models.UserDashboardModel{}, err2
	}
	fmt.Println(groupsArr)

	// TODO: Get all the chats from the groups
	var groupRecent []models.RecentChatModel
	for i := range groupsArr {
		data, err3 := u.groupMsg.GetAllMessagesFromGroup(groupsArr[i])
		if err3 != nil {
			return models.UserDashboardModel{}, err3
		}
		if len(data) == 0 {
			continue
		}

		for j := range data {
			val := models.RecentChatModel{
				UserName:    data[j].Name,
				UserAvatar:  data[j].Avatar,
				LastMsg:     data[j].Message,
				LastMsgTime: data[j].Time,
				UserPhone:   strconv.Itoa(groupsArr[j]),
			}

			groupRecent = append(groupRecent, val)
		}
	}
	fmt.Println("----------------------------------------------------------------------------")
	fmt.Println(groupRecent)

	// TODO: Delete the chats and keep only the last one
	// recentChatList - contains a list of all the chats from and to the current user
	var recent []models.MessageModel
	for i := range recentChatList {
		myTime, err := time.Parse("02-01-2006 3:04:05 PM", recentChatList[i].Time)
		if err != nil {
			return models.UserDashboardModel{}, err
		}
		diff := time.Now().Sub(myTime)
		d := models.MessageModel{
			From:    recentChatList[i].From,
			To:      recentChatList[i].To,
			Content: recentChatList[i].Content,
			Time:    recentChatList[i].Time,
			Status:  recentChatList[i].Status,
			Order:   float64(diff),
		}
		recent = append(recent, d)
	}

	sort.Slice(recent, func(i, j int) bool {
		return recent[i].Order > recent[j].Order
	})

	var (
		fromOrder, toOrder int
		recentChats        []models.RecentChatModel
	)

	if len(recent) == 0 {
		return models.UserDashboardModel{}, errors.New("length is zero")
	}
	for i := range recent {
		for j := i; j < len(recent); j++ {
			if recent[j].From == phone {
				fromOrder = j
				break
			}
		}

		for j := i; j < len(recent); j++ {
			if recent[j].To == phone {
				toOrder = j
				break
			}
		}

		// This only gives one chat result, we need a list of chat with different targets
		if fromOrder < toOrder {
			mdl, err4 := u.userData.GetUserData(recent[fromOrder].From)
			if err4 != nil {
				return models.UserDashboardModel{}, err4
			}
			fmt.Println("data of fromOrder in to - ", mdl)

			d := models.RecentChatModel{
				UserName:    mdl.UserName,
				UserAvatar:  mdl.UserAvatarUrl,
				UserPhone:   recent[fromOrder].To,
				LastMsg:     recent[fromOrder].Content,
				LastMsgTime: recent[fromOrder].Time,
			}
			recentChats = append(recentChats, d)
		} else {
			mdl, _ := u.userData.GetUserData(recent[toOrder].To)

			fmt.Println("data of toOrder in from - ", mdl)

			d := models.RecentChatModel{
				UserName:    mdl.UserName,
				UserAvatar:  mdl.UserAvatarUrl,
				UserPhone:   recent[toOrder].From,
				LastMsg:     recent[toOrder].Content,
				LastMsgTime: recent[toOrder].Time,
			}
			recentChats = append(recentChats, d)
		}
	}

	recentChats = append(recentChats, groupRecent...)

	fmt.Println("--------------- RECENT CHAT ARRAY -----------------")

	for i := 0; i < len(recentChats); i++ {
		for j := i; j < len(recentChats); j++ {
			if i == j {
				continue
			}
			if recentChats[i].UserPhone == recentChats[j].UserPhone {
				recentChats = append(recentChats[:i], recentChats[i+1:]...)
			}
		}
	}

	fmt.Println(recentChats)

	data := models.UserDashboardModel{
		UserPhone:      phone,
		UserDetails:    userDetails,
		RecentChatList: recentChats,
		StoryList:      nil,
	}

	return data, nil
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
