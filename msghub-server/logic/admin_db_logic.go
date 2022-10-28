package logic

import (
	"errors"
	"msghub-server/models"
	"msghub-server/repository"

	"gorm.io/gorm"
)

type AdminDb struct {
	repo repository.Admin
	err  error
}

// MigrateAdminDb :  Creates table for admin according the struct Admin
func MigrateAdminDb(db *gorm.DB) error {
	err := db.AutoMigrate(&repository.Admin{})
	return err
}

func (admin AdminDb) AdminLoginLogic(username, password string) error {
	data, err := admin.repo.LoginAdmin(username, password)
	if err != nil {
		return err
	}

	if data.AdminName == username {
		if data.AdminPass == password {
			return nil
		}
		return errors.New("you have entered wrong password, please try again")
	}
	return errors.New("you have entered wrong password, please try again")
}

func (admin AdminDb) GetAllAdminsData(name string) ([]models.AdminModel, error) {
	data, err := admin.repo.GetAdminsData(name)

	return data, err
}

func (admin AdminDb) GetUsersData() ([]models.UserModel, error) {
	data, err := admin.repo.GetAllUsersData()

	return data, err
}

func (admin AdminDb) GetGroupsData() ([]models.GroupModel, error) {
	data, err := admin.repo.GetGroupsData()

	return data, err
}

func (admin AdminDb) BlockThisUserLogic(id, condition string) error {
	err := admin.repo.AdminBlockThisUserRepo(id, condition)

	return err
}
