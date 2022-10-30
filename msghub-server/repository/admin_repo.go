package repository

import (
	"errors"
	"log"
	"msghub-server/models"
	"strconv"
)

type Admin struct {
	AdminId   int    `gorm:"not null;primaryKey;autoIncrement" json:"admin_id"`
	AdminName string `gorm:"not null" json:"admin_name"`
	AdminPass string `gorm:"not null" json:"admin_pass"`
}

func (admin Admin) LoginAdmin(uname, pass string) (Admin, error) {
	var name, password string
	rows, err := models.SqlDb.Query(
		`SELECT 
    	admin_name,
    	admin_pass
	FROM admins
	WHERE admin_name = $1;`, uname)

	if err != nil {
		return Admin{}, errors.New("an unknown error occurred, please try again")
	}

	defer rows.Close()
	for rows.Next() {
		if err1 := rows.Scan(
			&name,
			&password,
		); err1 != nil {
			return Admin{}, err1
		}
	}

	data := Admin{
		AdminName: name,
		AdminPass: password,
	}

	return data, nil
}

func (admin Admin) GetAdminsData(uname string) ([]models.AdminModel, error) {
	var (
		adminID, adminName string
		res                []models.AdminModel
	)
	rows, err := models.SqlDb.Query(
		`SELECT 
		admin_id, 
    	admin_name
	FROM admins
	WHERE admin_name != $1;`, uname)

	if err != nil {
		return res, errors.New("an unknown error occurred, please try again")
	}

	for rows.Next() {
		if err := rows.Scan(
			&adminID,
			&adminName,
		); err != nil {
			return res, err
		}

		data := models.AdminModel{
			AdminId:   adminID,
			AdminName: adminName,
		}

		res = append(res, data)
	}

	return res, nil
}

func (admin Admin) GetAllUsersData() ([]models.UserModel, error) {
	var (
		phone, name, about string
		avatar             *string
		isBlocked          bool
		res                []models.UserModel
	)
	rows, err := models.SqlDb.Query(
		`SELECT user_ph_no, user_name, user_avatar, user_about, is_blocked FROM users;`)
	if err != nil {
		return res, errors.New("an unknown error occurred, please try again")
	}

	for rows.Next() {
		if err := rows.Scan(
			&phone,
			&name,
			&avatar,
			&about,
			&isBlocked,
		); err != nil {
			return res, err
		}

		null := ""
		if avatar == nil {
			avatar = &null
		}

		data := models.UserModel{
			UserPhone:     phone,
			UserAvatarUrl: *avatar,
			UserName:      name,
			UserAbout:     about,
			UserBlocked:   isBlocked,
		}

		res = append(res, data)
	}

	return res, nil
}

func (admin Admin) GetGroupsData() ([]models.GroupModel, error) {
	var (
		id, name, about, date, members, creator string
		avatar                                  *string

		isBanned bool
		res      []models.GroupModel
	)
	rows, err := models.SqlDb.Query(
		`SELECT group_id, group_name, group_avatar, group_about, group_creator, group_created_date, group_total_members, is_banned FROM groups;`)
	if err != nil {
		return res, errors.New("an unknown error occurred, please try again")
	}

	for rows.Next() {
		if err := rows.Scan(
			&id,
			&name,
			&avatar,
			&about,
			&creator,
			&date,
			&members,
			&isBanned,
		); err != nil {
			return res, err
		}

		null := ""
		if avatar == nil {
			avatar = &null
		}

		m, err := strconv.Atoi(members)
		if err != nil {
			return res, err
		}

		data := models.GroupModel{
			Id:          id,
			Owner:       creator,
			Image:       *avatar,
			Name:        name,
			About:       about,
			CreatedDate: date,
			NoOfMembers: m,
			IsBanned:    isBanned,
		}

		res = append(res, data)
	}

	return res, nil
}

func (admin Admin) AdminBlockThisUserRepo(id, condition string) error {
	_, err1 := models.SqlDb.Exec(`UPDATE users SET is_blocked = true, block_duration = $1 WHERE user_ph_no = $2 AND is_blocked = false;`, condition, id)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (admin Admin) AdminBlockThisGroupRepo(id, condition string) error {
	_, err1 := models.SqlDb.Exec(`UPDATE groups SET is_banned = true, banned_time = $1 WHERE group_id = $2 AND is_banned = false;`, condition, id)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}
