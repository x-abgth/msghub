package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
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

var GormDb *gorm.DB
var SqlDb *sql.DB

func ConnectDb() {
	// loads env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(".env file loading error -- ", err)
		os.Exit(0)
	}

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

	var err1 error
	GormDb, err = gorm.Open(postgres.Open(psql), &gorm.Config{})
	SqlDb, err1 = sql.Open("postgres", psql)
	if err != nil {
		log.Fatal("Error connecting to database - ", err.Error())
	}
	if err1 != nil {
		log.Fatal("Error connecting to database without gorm - ", err1.Error())
	}
}
