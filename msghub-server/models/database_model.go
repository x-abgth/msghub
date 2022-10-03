package models

import (
	"database/sql"
	"gorm.io/gorm"
)

var (
	SqlDb  *sql.DB
	GormDb *gorm.DB
)
