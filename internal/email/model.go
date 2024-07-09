package email

import (
	"time"
)

type EmailRequest struct {
	ID              string    `json:"_id"`
	FromEmail       string    `json:"fromEmail"`
	Subject         string    `json:"subject"`
	Body            string    `json:"body"`
	FileLinks       []string  `json:"fileLinks"`
	scheduledSendAt time.Time `bson:"scheduledSendAt"`
}

type EmailRequestList struct {
	ID     string   `json:"_id"`
	ToList []string `json:"toList"`
}
