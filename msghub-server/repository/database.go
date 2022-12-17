package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/x-abgth/msghub/msghub-server/models"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type config struct {
	host    string
	port    string
	user    string
	pass    string
	dbName  string
	sslMode string
}

func ConnectDb() {
	// loads env file
	configure := &config{
		host:    os.Getenv("DB_HOST"),
		port:    os.Getenv("DB_PORT"),
		user:    os.Getenv("DB_USER"),
		pass:    os.Getenv("DB_PASS"),
		dbName:  os.Getenv("DB_NAME"),
		sslMode: os.Getenv("DB_SSLMODE"),
	}

	psql := fmt.Sprintf("host= %s port= %s user= %s password= %s dbname= %s sslmode= %s",
		configure.host,
		configure.port,
		configure.user,
		configure.pass,
		configure.dbName,
		configure.sslMode)

	var err1, err error
	models.GormDb, err = gorm.Open(postgres.Open(psql), &gorm.Config{})
	models.SqlDb, err1 = sql.Open("postgres", psql)
	if err != nil {
		log.Fatal("Error connecting to repository - ", err.Error())
	}
	if err1 != nil {
		log.Fatal("Error connecting to repository without gorm - ", err1.Error())
	}
}
