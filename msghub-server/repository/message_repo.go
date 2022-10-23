package repository

import (
	"errors"
	"fmt"
	"log"
	"msghub-server/models"
)

type Message struct {
	MsgId      int    `gorm:"not null;primaryKey;autoIncrement:true" json:"msg_id"`
	FromUserId string `gorm:"not null" json:"from_user_id"`
	ToUserId   string `gorm:"not null" json:"to_user_id"`
	Content    string `gorm:"not null" json:"content"`
	SentTime   string `gorm:"not null" json:"sent_time"`
	Status     string `gorm:"not null" json:"status"`
}

func (m Message) InsertMessageDataRepository(data Message) error {
	fmt.Println("In repo = ", data)

	_, err1 := models.SqlDb.Exec(`INSERT INTO messages(from_user_id, to_user_id, content, sent_time, status) 
VALUES($1, $2, $3, $4, $5);`,
		data.FromUserId, data.ToUserId, data.Content, data.SentTime, data.Status)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (m Message) GetAllPersonalMessages(from, to string) ([]models.MessageModel, error) {

	var (
		fromID, msg, time string
		res               []models.MessageModel
	)

	rows, err := models.SqlDb.Query(
		`SELECT 
    	from_user_id, 
    	content,
    	sent_time
	FROM messages
	WHERE from_user_id = $1 AND to_user_id = $2;`, from, to)
	if err != nil {
		return res, err
	}

	defer rows.Close()

	for rows.Next() {
		if err1 := rows.Scan(
			&fromID,
			&msg,
			&time); err1 != nil {
			return res, err1
		}

		data := models.MessageModel{
			From:    fromID,
			Content: msg,
			Time:    time,
		}
		res = append(res, data)
	}

	return res, nil
}
