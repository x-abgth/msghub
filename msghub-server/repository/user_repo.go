package repository

import (
	"errors"
	"log"
	"msghub-server/models"
	"strconv"
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

func (user User) RegisterUser(formName, formPhone, formPass string) (bool, error) {
	defaultAbout := "Hey there! Send me a Hi."

	_, err1 := models.SqlDb.Exec(`INSERT INTO users(user_name, user_about, user_ph_no, user_password, is_blocked) 
VALUES($1, $2, $3, $4, $5);`,
		formName, defaultAbout, formPhone, formPass, false)
	if err1 != nil {
		log.Println(err1)
		return false, errors.New("sorry, An unknown error occurred. Please try again")
	}

	return true, nil
}

func (user User) UserDuplicationStatus(phone string) int {
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

func (user User) GetUserData(ph string) (models.UserModel, error) {
	var name, phone, about, isBlocked string
	var avatar *string
	var data models.UserModel

	rows, err := models.SqlDb.Query(
		`SELECT 
    	user_avatar, 
    	user_name, 
    	user_ph_no,
    	user_about,
    	is_blocked
	FROM users
	WHERE user_ph_no = $1 AND is_blocked = $2;`, ph, false)
	if err != nil {
		return data, err
	}

	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
		if err1 := rows.Scan(
			&avatar,
			&name,
			&phone,
			&about,
			&isBlocked); err1 != nil {
			return data, err1
		}
	}

	if avatar == nil {
		null := ""

		avatar = &null
	}
	data = models.UserModel{
		UserAvatarUrl: *avatar,
		UserName:      name,
		UserPhone:     phone,
		UserAbout:     about,
		UserBlocked:   isBlocked,
	}

	return data, nil
}

// This is not actual list, need to update
func (user User) GetRecentChatList(ph string) ([]models.MessageModel, error) {
	var from, to, content, sentTime, status string

	var res []models.MessageModel

	rows, err := models.SqlDb.Query(
		`SELECT 
    from_user_id,
    	to_user_id, 
    	content, 
    	sent_time,
    	status
	FROM messages
	WHERE from_user_id = $1 OR to_user_id = $2`, ph, ph)
	if err != nil {
		return res, err
	}

	defer rows.Close()

	for rows.Next() {
		if err1 := rows.Scan(
			&from,
			&to,
			&content,
			&sentTime,
			&status); err1 != nil {
			return res, err1
		}

		data := models.MessageModel{
			From:    from,
			To:      to,
			Content: content,
			Time:    sentTime,
			Status:  status,
		}

		res = append(res, data)
	}

	return res, nil
}

func (user User) GetAllUsersData(ph string) ([]models.UserModel, error) {
	var name, phone, about string
	var avatar *string
	var res []models.UserModel

	rows, err := models.SqlDb.Query(
		`SELECT 
    	user_avatar, 
    	user_name, 
    	user_about,
    	user_ph_no 
	FROM users
	WHERE is_blocked = $1 AND user_ph_no != $2;`, false, ph)
	if err != nil {
		return res, err
	}

	defer rows.Close()

	for rows.Next() {
		if err1 := rows.Scan(
			&avatar,
			&name,
			&about,
			&phone); err1 != nil {
			return res, err1
		}

		if avatar == nil {
			null := ""

			avatar = &null
		}
		data := models.UserModel{
			UserName:      name,
			UserPhone:     phone,
			UserAbout:     about,
			UserAvatarUrl: *avatar,
		}
		res = append(res, data)
	}

	return res, nil
}

func (user User) GetGroupForUser(userId string) ([]int, error) {
	var group, userPh, role string
	rows, err := models.SqlDb.Query(
		`SELECT 
    	group_id, 
    	user_id, 
    	user_role
	FROM user_group_relations
	WHERE user_id = $1;`, userId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var res []int
	for rows.Next() {
		if err1 := rows.Scan(
			&group,
			&userPh,
			&role); err1 != nil {
			return nil, err1
		}
		n, _ := strconv.Atoi(group)
		res = append(res, n)
	}
	return res, nil
}

func (user User) UpdateUserData(model models.UserModel) error {
	_, err1 := models.SqlDb.Exec(`UPDATE users SET user_name = $1, user_about = $2, user_avatar = $3 WHERE user_ph_no = $4;`,
		model.UserName, model.UserAbout, model.UserAvatarUrl, model.UserPhone)
	if err1 != nil {
		return errors.New("couldn't execute the sql query")
	}

	return nil
}
