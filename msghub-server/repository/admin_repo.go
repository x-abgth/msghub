package repository

import (
	"errors"
	"msghub-server/models"
)

type Admin struct {
	AdminId   int    `gorm:"not null;primaryKey;autoIncrement" json:"admin_id"`
	AdminName string `gorm:"not null" json:"admin_name"`
	AdminPass string `gorm:"not null" json:"admin_pass"`
}

func (admin Admin) LoginAdmin(uname, pass string) (bool, error) {
	rows, err := models.SqlDb.Query(
		`SELECT 
    	admin_name,
    	admin_pass
	FROM admins
	WHERE admin_name = $1;`, uname)

	if err != nil {
		return false, errors.New("an unknown error occurred, please try again")
	}

	defer rows.Close()
	for rows.Next() {
		if err1 := rows.Scan(
			&admin.AdminName,
			&admin.AdminPass,
		); err1 != nil {
			return false, err1
		}
	}

	// TODO: Create a table and add a dummy admin user
	// TODO: Initialize admin data to a model class to pass it to the logic module
	return true, nil
}
