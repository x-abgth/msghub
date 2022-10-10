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
