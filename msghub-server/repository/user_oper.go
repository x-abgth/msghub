package repository

import (
	"log"
	"msghub-server/models"
)

type User struct {
	UserPhNo     string  `gorm:"not null;primaryKey;autoIncrement:false" json:"user_ph_no"`
	UserName     string  `gorm:"not null" json:"user_name"`
	UserAvatar   *string `json:"user_avatar"`
	UserAbout    string  `gorm:"not null" json:"user_about"`
	UserPassword string  `gorm:"not null" json:"user_password"`
	IsBlocked    bool    `gorm:"not null" json:"is_blocked"`
}

func (user User) GetUserDataUsingPhone(formPhone string) (int, User, error) {

	rows, err := models.SqlDb.Query(
		`SELECT 
    	user_avatar, 
    	user_about,
    	user_name, 
    	user_ph_no,
    	user_password, 
    	is_blocked
	FROM users
	WHERE user_ph_no = $1;`, formPhone)
	if err != nil {
		return 0, user, err
	}

	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
		if err1 := rows.Scan(
			&user.UserAvatar,
			&user.UserAbout,
			&user.UserName,
			&user.UserPhNo,
			&user.UserPassword,
			&user.IsBlocked); err1 != nil {
			return 0, user, err1
		}
	}

	return count, user, nil
}

func RegisterUser(formName, formPhone, formPass string) (bool, string) {
	defaultAbout := "Hey there! Send me a Hi."

	_, err1 := models.SqlDb.Exec(`INSERT INTO users(user_name, user_about, user_ph_no, user_password, is_blocked) 
VALUES($1, $2, $3, $4, $5);`,
		formName, defaultAbout, formPhone, formPass, false)
	if err1 != nil {
		log.Fatal(err1)
		return false, "Sorry, An unknown error occurred. Please try again."
	}

	return true, ""
}

func UserDuplicationStatus(phone string) int {
	var total = 0

	rows, err := models.SqlDb.Query(
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
	var name, phone, isBlocked string
	var avatar *string
	rows, err := models.SqlDb.Query(
		`SELECT 
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
		UserAvatarUrl: *avatar,
		UserName:      name,
		UserPhone:     phone,
		UserBlocked:   isBlocked,
	}

	return data
}

// This is not actual list, need to update
func GetRecentChatList(ph string) []models.RecentChatModel {
	var name, phone string
	var avatar *string
	rows, err := models.SqlDb.Query(
		`SELECT 
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
