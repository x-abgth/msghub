package repository

import (
	"errors"
	"log"
	"msghub-server/models"
)

type Group struct {
	GroupId           int    `gorm:"not null;primaryKey;autoIncrement:true" json:"group_id"`
	GroupName         string `gorm:"not null" json:"group_name"`
	GroupAvatar       string `gorm:"not null" json:"group_avatar"`
	GroupAbout        string `gorm:"not null" json:"group_about"`
	GroupCreator      string `gorm:"not null" json:"group_creator"`
	GroupCreatedDate  string `gorm:"not null" json:"group_created_date"`
	GroupTotalMembers int    `gorm:"not null" json:"group_total_members"`
	IsBanned          bool   `gorm:"not null" json:"is_banned"`
}

type UserGroupRelation struct {
	Id       int    `gorm:"not null;primaryKey;autoIncrement:true" json:"id"`
	GroupId  int    `gorm:"not null" json:"group_id"`
	UserId   string `gorm:"not null" json:"user_id"`
	UserRole string `gorm:"not null" json:"user_role"`
}

type GroupMessage struct {
	MsgId          int    `gorm:"not null;primary key;autoIncrement:true" json:"msg_id"`
	GroupId        int    `gorm:"not null" json:"group_id"`
	SenderId       string `gorm:"not null" json:"sender_id"`
	MessageContent string `gorm:"not null" json:"message_content"`
	SentTime       string `gorm:"not null" json:"sent_time"`
}

func CreateGroup(data Group) (int, error) {
	var id int
	if err := models.SqlDb.QueryRow(`INSERT INTO groups
		(group_name, group_avatar, group_about, group_creator, group_created_date, group_total_members, is_banned) 
VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING group_id`, data.GroupName, data.GroupAvatar, data.GroupAbout, data.GroupCreator, data.GroupCreatedDate, data.GroupTotalMembers, data.IsBanned).Scan(&id); err != nil {
		log.Println(err.Error())
		return 0, errors.New("sorry, An unknown error occurred. Please try again")
	}

	return id, nil
}

func (relation UserGroupRelation) CreateUserGroupRelation(groupId int, userId, role string) error {
	_, err1 := models.SqlDb.Exec(`INSERT INTO user_group_relations(
	                 group_id, user_id, user_role)
	VALUES($1, $2, $3);`,
		groupId, userId, role)
	if err1 != nil {
		log.Println(err1.Error())
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (gm GroupMessage) InsertGroupMessagesRepo(message GroupMessage) error {
	_, err1 := models.SqlDb.Exec(`INSERT INTO group_messages(
	                 group_id, sender_id, message_content, sent_time)
	VALUES($1, $2, $3, $4);`,
		message.GroupId, message.SenderId, message.MessageContent, message.SentTime)
	if err1 != nil {
		log.Println(err1.Error())
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (relation UserGroupRelation) GetAllGroupsForAUser(ph string) ([]int, error) {
	var (
		num int
		res []int
	)
	rows, err := models.SqlDb.Query(
		`SELECT 
    	group_id
	FROM user_group_relations
	WHERE user_id = $1;`, ph)
	if err != nil {
		return res, err
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&num); err != nil {
			return res, err
		}
		res = append(res, num)
	}

	return res, nil
}

func (gm GroupMessage) GetAllMessagesFromGroup(id int) ([]models.GrpMsgModel, error) {
	var (
		name, avtr, sender, content, time string
		res                               []models.GrpMsgModel
	)
	rows, err := models.SqlDb.Query(
		`SELECT groups.group_name, groups.group_avatar, group_messages.sender_id, group_messages.message_content, group_messages.sent_time 
FROM groups 
    INNER JOIN group_messages 
        ON groups.group_id = group_messages.group_id WHERE groups.group_id = $1;`, id)
	if err != nil {
		return res, err
	}

	defer rows.Close()

	for rows.Next() {
		if err1 := rows.Scan(
			&name,
			&avtr,
			&sender,
			&content,
			&time,
		); err1 != nil {
			return res, err1
		}

		data := models.GrpMsgModel{
			Name:    name,
			Avatar:  avtr,
			Sender:  sender,
			Message: content,
			Time:    time,
		}

		res = append(res, data)
	}
	return res, nil

}
