package database

import (
	"log"
	"msghub-server/models"
	"msghub-server/utils"

	"gorm.io/gorm"
)

type User struct {
	UserID       int64  `gorm:"primaryKey;autoIncrement" json:"user_id"`
	UserAvatar   string `json:"user_avatar"`
	UserName     string `json:"user_name"`
	UserPhNo     string `json:"user_ph_no"`
	UserPassword string `json:"user_password"`
	IsBlocked    bool   `json:"is_blocked"`
}

func MigrateUser(db *gorm.DB) error {
	err := db.AutoMigrate(&User{})
	return err
}

func LoginUserWithCredentials(formPhone, formPassword string) (bool, string) {
	var name, phone, pass, dialog string
	var avatar *string
	var isBlocked bool

	var flag bool
	rows, err := SqlDb.Query(
		`SELECT 
    	user_avatar, 
    	user_name, 
    	user_ph_no,
    	user_password, 
    	is_blocked
	FROM users
	WHERE user_ph_no = $1;`, formPhone)
	if err != nil {
		flag = false
		log.Fatal("Error - ", err)
	}

	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
		if err1 := rows.Scan(
			&avatar,
			&name,
			&phone,
			&pass,
			&isBlocked); err1 != nil {
			flag = false
			log.Fatal("Error - ", err1)
		}
	}

	// Check the value is isBlocked and if string convert to bool using if
	if isBlocked {
		flag = false
		dialog = "You are temporarily blocked from this application!"
	} else if count < 1 {
		flag = false
		dialog = "You don't have an account, Please register."
	} else if count > 1 {
		flag = false
		dialog = "Something went wrong. Try login again!"
		// SHOULD DELETE EXTRA REGISTERED NUMBER!
	} else {
		if phone == formPhone {
			if utils.CheckPasswordMatch(formPassword, pass) {
				flag = true
				var user models.UserModel

				var blank = ""
				if avatar == nil {
					avatar = &blank
				}
				user = models.UserModel{
					UserAvatarUrl: *avatar,
					UserName:      name,
					UserPhone:     phone,
				}

				models.InitUserModel(user)
			} else {
				flag = false
				dialog = "Invalid phone number or password!"
			}
		} else {
			flag = false
			dialog = "Invalid phone number or password!"
		}
	}

	return flag, dialog
}

func RegisterUser(formName, formPhone, formPass string) (bool, string) {

	_, err1 := SqlDb.Exec(`INSERT INTO users(user_name, user_ph_no, user_password, is_blocked) 
VALUES($1, $2, $3, $4);`,
		formName, formPhone, formPass, false)
	if err1 != nil {
		log.Fatal(err1)
		return false, "Sorry, An unknown error occured. Please try again."
	}

	return true, ""
}

func UserDuplicationStatus(phone string) int {
	var total = 0

	rows, err := SqlDb.Query(
		`SELECT *
	FROM users
	WHERE user_ph_no = $1;`, phone)
	if err != nil {
		log.Fatal("Error - ", err)
	}

	defer rows.Close()
	for rows.Next() {
		total++
	}

	return total
}

func GetUserData(ph string) models.UserModel {
	var id, name, phone, isBlocked string
	var avatar *string
	rows, err := SqlDb.Query(
		`SELECT 
		user_id,
    	user_avatar, 
    	user_name, 
    	user_ph_no,
    	is_blocked
	FROM users
	WHERE user_ph_no = $1 AND is_blocked = $2;`, ph, false)
	if err != nil {
		log.Fatal("Error - ", err)
	}

	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
		if err1 := rows.Scan(
			&id,
			&avatar,
			&name,
			&phone,
			&isBlocked); err1 != nil {
			log.Fatal("Error - ", err1)
		}
	}

	if avatar == nil {
		null := ""

		avatar = &null
	}
	data := models.UserModel{
		UserID:        id,
		UserAvatarUrl: *avatar,
		UserName:      name,
		UserPhone:     phone,
		UserBlocked:   isBlocked,
	}

	return data
}

// This is not actual list, need to update
func GetRecentChatList(ph string) []models.RecentChatModel {
	var id, name, phone string
	var avatar *string
	rows, err := SqlDb.Query(
		`SELECT 
		user_id,
    	user_avatar, 
    	user_name, 
    	user_ph_no 
	FROM users
	WHERE is_blocked = $1 AND user_ph_no != $2;`, false, ph)
	if err != nil {
		log.Fatal("Error - ", err)
	}

	defer rows.Close()

	var res []models.RecentChatModel
	for rows.Next() {
		if err1 := rows.Scan(
			&id,
			&avatar,
			&name,
			&phone); err1 != nil {
			log.Fatal("Error - ", err1)
		}

		if avatar == nil {
			null := ""

			avatar = &null
		}
		data := models.RecentChatModel{
			UserName:    name,
			UserPhone:   phone,
			UserAvatar:  *avatar,
			LastMsg:     "Hi",
			LastMsgTime: "NIL",
		}
		res = append(res, data)
	}

	return res
}
