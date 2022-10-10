package logic

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"msghub-server/models"
	"msghub-server/repository"
	"time"
)

type GroupDataLogicModel struct {
	groupTb           repository.Group
	userGroupRelation repository.UserGroupRelation
	err               error
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

func (group GroupDataLogicModel) CreateGroupAndInsertDataLogic(groupData models.GroupModel) (bool, error) {
	// Get date of the group created
	t := time.Now()
	dateOfCreation := fmt.Sprintf("%s", t.Format("02/01/2006"))

	data := repository.Group{
		GroupName:         groupData.Name,
		GroupAvatar:       groupData.Image,
		GroupAbout:        groupData.About,
		GroupCreator:      groupData.Owner,
		GroupCreatedDate:  dateOfCreation,
		GroupTotalMembers: len(groupData.Members),
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

	return true, nil
}
