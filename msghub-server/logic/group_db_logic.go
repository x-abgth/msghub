package logic

import (
	"log"
	"msghub-server/models"
	"msghub-server/repository"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type GroupDataLogicModel struct {
	groupTb           repository.Group
	userGroupRelation repository.UserGroupRelation
	messageGroupTb    repository.GroupMessage
}

// MigrateUserDb :  Creates table for user according the struct User
func (group GroupDataLogicModel) MigrateGroupDb(db *gorm.DB) error {
	err := db.AutoMigrate(&repository.Group{})
	return err
}

func (group GroupDataLogicModel) MigrateUserGroupDb(db *gorm.DB) error {
	err := db.AutoMigrate(&repository.UserGroupRelation{})
	return err
}

func (group GroupDataLogicModel) MigrateGroupMessagesDb(db *gorm.DB) error {
	err := db.AutoMigrate(&repository.GroupMessage{})
	return err
}

func (group GroupDataLogicModel) CreateGroupAndInsertDataLogic(groupData models.GroupModel) (bool, error) {
	// Get date of the group created
	t := time.Now()
	dateOfCreation := t.Format("02/01/2006")

	data := repository.Group{
		GroupName:         groupData.Name,
		GroupAvatar:       groupData.Image,
		GroupAbout:        groupData.About,
		GroupCreator:      groupData.Owner,
		GroupCreatedDate:  dateOfCreation,
		GroupTotalMembers: len(groupData.Members) + 1,
		IsBanned:          false,
	}

	id, err := repository.CreateGroup(data)
	if err != nil {
		log.Println(err.Error())
		return false, err
	}

	err1 := group.userGroupRelation.CreateUserGroupRelation(id, groupData.Owner, "admin")
	if err != nil {
		log.Println(err1.Error())
		return false, err1
	}
	for i := range groupData.Members {
		err := group.userGroupRelation.CreateUserGroupRelation(id, groupData.Members[i], "member")
		if err != nil {
			log.Println(err.Error())
			return false, err
		}
	}

	msg := models.GroupMessageModel{
		GroupId:  strconv.Itoa(id),
		SenderId: "admin",
		Content:  "+91 " + groupData.Owner + " created a group named " + groupData.Name + ".",
		Time:     time.Now().Format("02-01-2006 3:04:05 PM"),
	}
	err2 := group.InsertMessagesToGroup(msg)
	if err2 != nil {
		return false, err2
	}

	return true, nil
}

func (group GroupDataLogicModel) InsertMessagesToGroup(message models.GroupMessageModel) error {
	var (
		err error
	)

	group.messageGroupTb.GroupId, err = strconv.Atoi(message.GroupId)
	if err != nil {
		return err
	}
	group.messageGroupTb.SenderId = message.SenderId
	group.messageGroupTb.MessageContent = message.Content
	group.messageGroupTb.SentTime = message.Time

	err1 := group.messageGroupTb.InsertGroupMessagesRepo(group.messageGroupTb)
	if err1 != nil {
		return err1
	}

	return nil
}