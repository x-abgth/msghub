package logic

import (
	"gorm.io/gorm"
	"msghub-server/repository"
)

type AdminDb struct {
	err error
}

// MigrateAdminDb :  Creates table for admin according the struct Admin
func MigrateAdminDb(db *gorm.DB) error {
	err := db.AutoMigrate(&repository.User{})
	return err
}
