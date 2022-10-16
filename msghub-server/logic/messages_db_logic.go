package logic

import (
	"gorm.io/gorm"
	"log"
	"msghub-server/models"
	"msghub-server/repository"
	"os"
	"sort"
	"time"
)

type MessageDb struct {
	UserData repository.Message
	err      error
}

// message status constants
const (
	IS_NOT_SENT  = "NOT_SENT"
	IS_SENT      = "SENT"
	IS_DELIVERED = "DELIVERED"
	IS_READ      = "READ"
)

// MigrateMessagesDb : Creates message table
func (m MessageDb) MigrateMessagesDb(db *gorm.DB) error {
	err := db.AutoMigrate(&repository.Message{})
	return err
}

func (m MessageDb) StorePersonalMessagesLogic(message repository.Message) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			os.Exit(1)
		}
	}()
	err := m.UserData.InsertMessageDataRepository(message)
	if err != nil {
		panic(err.Error())
	}
}

func (m MessageDb) GetMessageDataLogic(target string) ([]models.MessageModel, error) {
	var this []models.MessageModel

	err, data := m.UserData.GetAllPersonalMessages(target)
	if err != nil {
		return this, err
	}

	for i := range data {
		myTime, err := time.Parse("02-01-2006 3:04 PM", data[i].Time)
		if err != nil {
			return this, err
		}
		diff := time.Now().Sub(myTime)
		d := models.MessageModel{
			From:    data[i].From,
			Content: data[i].Content,
			Time:    data[i].Time,
			Order:   float64(diff),
		}

		this = append(this, d)
	}

	sort.Slice(this, func(i, j int) bool {
		return this[i].Order < this[i].Order
	})

	return this, nil
}
