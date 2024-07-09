package email

import (
	"context"
	"log"
	"time"

	template_package "newsletter-app/internal/template"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type EmailService struct {
	client *mongo.Client
}

func NewEmailService(client *mongo.Client) *EmailService {
	return &EmailService{client: client}
}

func (s *EmailService) ScheduleEmails() {

	for {

		now := time.Now()
		collection := s.client.Database("emailDB").Collection("templates")
		cur, err := collection.Find(context.TODO(), bson.M{"scheduled_send_at": bson.M{"$lte": now}})
		if err != nil {
			log.Println("Error finding scheduled emails:", err)
			continue
		}

		var templates []template_package.Template
		if err := cur.All(context.TODO(), &templates); err != nil {
			log.Println("Error decoding scheduled emails:", err)
			continue
		}

		for _, template := range templates {

			var emailRequest EmailRequest
			emailRequest.Body = template.Body
			emailRequest.FromEmail = template.FromEmail
			emailRequest.Subject = template.Subject
			emailRequest.FileLinks = template.FileLinks
			SendEmailSchedule(emailRequest)

			collection.UpdateOne(context.TODO(), bson.M{"_id": template.ID}, bson.M{"$set": bson.M{"scheduled_send_at": nil}})
		}

		time.Sleep(5 * time.Minute)
	}
}
