package repository

import (
	"errors"
	"log"
	"msghub-server/models"
	"strconv"
)

type User struct {
	UserPhNo        string  `gorm:"not null;primaryKey;autoIncrement:false" json:"user_ph_no"`
	UserName        string  `gorm:"not null" json:"user_name"`
	UserAvatar      *string `json:"user_avatar"`
	UserAbout       string  `gorm:"not null" json:"user_about"`
	UserPassword    string  `gorm:"not null" json:"user_password"`
	IsBlocked       bool    `gorm:"not null" json:"is_blocked"`
	BlockedDuration *string `json:"block_duration"`
	BlockList       *string `json:"block_list"`
}

type Storie struct {
	UserId          string `gorm:"primary key;not null;autoIncrement:false" json:"user_id"`
	StoryUrl        string `gorm:"not null" json:"story_url"`
	StoryUpdateTime string `gorm:"not null" json:"story_update_time"`
	Viewers         string `gorm:"not null" json:"viewers"`
	IsActive        bool   `gorm:"not null" json:"is_active"`
}

func (user User) GetUserDataUsingPhone(formPhone string) (int, models.UserModel, error) {
	var (
		model     models.UserModel
		userModel User
	)

	rows, err := models.SqlDb.Query(
		`SELECT 
    	user_avatar, 
    	user_about,
    	user_name, 
    	user_ph_no,
    	user_password, 
    	is_blocked, 
    	block_duration
	FROM users
	WHERE user_ph_no = $1;`, formPhone)
	if err != nil {
		return 0, models.UserModel{}, err
	}

	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
		if err1 := rows.Scan(
			&userModel.UserAvatar,
			&userModel.UserAbout,
			&userModel.UserName,
			&userModel.UserPhNo,
			&userModel.UserPassword,
			&userModel.IsBlocked,
			&userModel.BlockedDuration,
		); err1 != nil {
			return 0, models.UserModel{}, err1
		}
	}

	var blank = ""
	if userModel.UserAvatar == nil {
		userModel.UserAvatar = &blank
	}

	if userModel.BlockedDuration == nil {
		userModel.BlockedDuration = &blank
	}

	model = models.UserModel{
		UserAvatarUrl: *userModel.UserAvatar,
		UserAbout:     userModel.UserAbout,
		UserName:      userModel.UserName,
		UserPhone:     userModel.UserPhNo,
		UserPass:      userModel.UserPassword,
		UserBlocked:   userModel.IsBlocked,
		BlockDur:      *userModel.BlockedDuration,
	}

	return count, model, nil
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
	var name, phone, about string
	var isBlocked bool
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
	WHERE user_ph_no = $1;`, ph)
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
	WHERE is_recent = $1 AND (from_user_id = $2 OR to_user_id = $3) ORDER BY sent_time;`, true, ph, ph)
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

func (user User) AddStoryRepo(model Storie) error {
	_, err1 := models.SqlDb.Exec(`INSERT INTO stories(user_id, story_url, story_update_time, viewers, is_active) 
VALUES($1, $2, $3, $4, $5);`, model.UserId, model.StoryUrl, model.StoryUpdateTime, model.Viewers, model.IsActive)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (user User) CheckUserStory(userId string) (bool, int) {
	var (
		status bool
		count  int
	)
	rows, err := models.SqlDb.Query(
		`SELECT 
    	is_active
	FROM stories
	WHERE user_id = $1;`, userId)
	if err != nil {
		return false, 0
	}

	defer rows.Close()

	for rows.Next() {
		count++
		if err1 := rows.Scan(
			&status); err1 != nil {
			return false, 0
		}
	}

	return status, count
}

func (user User) UpdateStoryStatusRepo(url, time, uid string) error {
	_, err1 := models.SqlDb.Exec(`UPDATE stories SET story_url = $1, story_update_time = $2, is_active = $3 WHERE user_id = $4;`, url, time, true, uid)
	if err1 != nil {
		return errors.New("couldn't execute the sql query")
	}

	return nil
}

func (user User) GetAllUserStories() []Storie {
	var (
		res                    []Storie
		id, url, time, viewers string
	)

	rows, err := models.SqlDb.Query(
		`SELECT 
    	user_id, story_url, story_update_time, viewers 
	FROM stories
	WHERE is_active = $1;`, true)
	if err != nil {
		return nil
	}

	defer rows.Close()

	for rows.Next() {
		if err1 := rows.Scan(&id, &url, &time, &viewers); err1 != nil {
			return nil
		}

		data := Storie{
			UserId:          id,
			StoryUrl:        url,
			StoryUpdateTime: time,
			Viewers:         viewers,
		}

		res = append(res, data)
	}

	return res
}

func (user User) UpdateUserData(model models.UserModel) error {
	_, err1 := models.SqlDb.Exec(`UPDATE users SET user_name = $1, user_about = $2, user_avatar = $3 WHERE user_ph_no = $4;`,
		model.UserName, model.UserAbout, model.UserAvatarUrl, model.UserPhone)
	if err1 != nil {
		return errors.New("couldn't execute the sql query")
	}

	return nil
}

func (user User) UndoAdminBlockRepo(id string) error {
	_, err1 := models.SqlDb.Exec(`UPDATE users SET is_blocked = false, block_duration = '' WHERE user_ph_no = $1;`, id)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (user User) UnblockGroupRepo(id string) error {
	_, err1 := models.SqlDb.Exec(`UPDATE groups SET is_banned = false, banned_time = '' WHERE group_id = $1;`, id)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (user User) GetUserBlockList(id string) (string, error) {
	var blockList *string
	rows, err := models.SqlDb.Query(
		`SELECT 
    	block_list
	FROM users
	WHERE user_ph_no = $1;`, id)
	if err != nil {
		return "", err
	}

	defer rows.Close()

	for rows.Next() {
		if err1 := rows.Scan(
			&blockList); err1 != nil {
			return "", err1
		}
	}

	null := ""
	if blockList == nil {
		blockList = &null
	}

	return *blockList, nil
}

func (user User) UpdateUserBlockList(id, val string) error {
	_, err1 := models.SqlDb.Exec(`UPDATE users SET block_list = $1 WHERE user_ph_no = $2;`, val, id)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}
