package template

import (
	"time"
)

type Template struct {
	ID              string     `bson:"_id,omitempty"`
	FromEmail       string     `bson:"fromEmail"`
	Subject         string     `bson:"subject"`
	Body            string     `bson:"body"`
	FileLinks       []string   `bson:"fileLinks"`
	ScheduledSendAt *time.Time `bson:"scheduledSendAt"`
}
