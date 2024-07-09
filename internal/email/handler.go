package email

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	//"os"
	"github.com/joho/godotenv"

	recipent "newsletter-app/internal/recipient"
	"newsletter-app/pkg/email"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Handler struct {
	client *mongo.Client
}

func NewHandler(client *mongo.Client) *Handler {
	return &Handler{client: client}
}

func SendEmailSchedule(emailRequest EmailRequest) bool {
	var h Handler

	collection := h.client.Database("emailDB").Collection("recipients")
	cur, err := collection.Find(context.TODO(), bson.M{"unsubscribed": false})
	if err != nil {
		return false
	}
	defer cur.Close(context.TODO())

	var recipients []string
	for cur.Next(context.TODO()) {
		var recipient recipent.Recipient
		if err := cur.Decode(&recipient); err != nil {
			return false
		}
		recipients = append(recipients, recipient.Email)
	}

	for _, to := range recipients {
		attachments := downloadFilesFromS3(emailRequest.FileLinks)
		if err := email.SendEmail(emailRequest.FromEmail, to, emailRequest.Subject, emailRequest.Body, attachments); err != nil {

			return false
		}
	}
	return true

	// attachments := downloadFilesFromS3(emailRequest.FileLinks)

	// if err := email.SendEmail(emailRequest.From, emailRequest.To, emailRequest.Subject, emailRequest.Body, attachments); err != nil {

	// 	return true
	// }

}
func (h *Handler) SendEmailToList(w http.ResponseWriter, r *http.Request) {
	var emailRequestList EmailRequestList
	var emailRequest EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&emailRequestList); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	println("ToList", emailRequestList.ToList)
	//attachments := downloadFilesFromS3(emailRequest.FileLinks)
	collectionTemplates := h.client.Database("emailDB").Collection("templates")
	err := collectionTemplates.FindOne(context.TODO(), bson.M{"_id": emailRequestList.ID}).Decode(&emailRequest)
	if err != nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	println("->", emailRequest.FromEmail, emailRequest.Subject, emailRequest.Body)

	for _, to := range emailRequestList.ToList {
		println("to", to)

		attachments := downloadFilesFromS3(emailRequest.FileLinks)
		if err := email.SendEmail(
			emailRequest.FromEmail,
			to,
			emailRequest.Subject,
			emailRequest.Body,
			attachments,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	fmt.Fprintf(w, "Email sent successfully")
}

func (h *Handler) SendEmail(w http.ResponseWriter, r *http.Request) {
	var emailRequest EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&emailRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	collectionTemplates := h.client.Database("emailDB").Collection("templates")
	err := collectionTemplates.FindOne(context.TODO(), bson.M{"_id": emailRequest.ID}).Decode(&emailRequest)
	print(emailRequest.ID)

	if err != nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	println("->", emailRequest.FromEmail, emailRequest.Subject, emailRequest.Body)

	collection := h.client.Database("emailDB").Collection("recipients")
	cur, err := collection.Find(context.TODO(), bson.M{"unsubscribed": false, "templateId": emailRequest.ID})
	if err != nil {
		http.Error(w, "Error ak otener los destinos", http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.TODO())

	var recipients []string
	for cur.Next(context.TODO()) {
		var recipient recipent.Recipient
		if err := cur.Decode(&recipient); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		recipients = append(recipients, recipient.Email)
	}
	println("emails_f", emailRequest.FileLinks[0])

	for _, to := range recipients {
		println("to", to)
		var body = emailRequest.Body + "<br> Desuscribrirte de este email: http://localhost:8081/unsubscribe/" + emailRequest.ID + "/" + to

		attachments := downloadFilesFromS3(emailRequest.FileLinks)
		if err := email.SendEmail(emailRequest.FromEmail, to, emailRequest.Subject, body, attachments); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	fmt.Fprintf(w, "Emails sent successfully")
}

// func (s *Handler) GetTemplate(id string) (*EmailRequest, error) {

// 	var template EmailRequest
// 	collection := s.client.Database("emailDB").Collection("templates")
// 	err = collection.FindOne(context.TODO(), bson.M{"_id": template.ID}).Decode(&template)
// 	if err != nil {
// 		return nil, err
// 	}

//		return &template, nil
//	}
func (h *Handler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	templateID := mux.Vars(r)["templateID"]
	email := mux.Vars(r)["email"]
	collection := h.client.Database("emailDB").Collection("recipients")
	update := bson.M{
		"$set": bson.M{
			"unsubscribed": true,
		},
	}
	_, err := collection.UpdateOne(context.TODO(), bson.M{"templateId": templateID, "email": email}, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	println("data", templateID, email)

	fmt.Fprintf(w, "Unsubscribed successfully")
}

func downloadFilesFromS3(links []string) []string {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error cargando archivo .env: %v", err)
	}

	awsAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-2"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsAccessKeyID, awsSecretAccessKey, "")),
	)

	//cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-west-2"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	var localPaths []string

	for _, link := range links {
		// Assume link format is "s3://bucket/key"
		println("link", link)
		bucket, key, err := parseS3Link(link)
		if err != nil {
			continue
		}
		println(bucket, "|", key)
		output, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			log.Printf("Failed to download file from S3: %v", err)
			continue
		}
		defer output.Body.Close()

		body, err := ioutil.ReadAll(output.Body)
		if err != nil {
			log.Printf("Failed to read file body: %v", err)
			continue
		}

		localPath := filepath.Join("/tmp", key)
		err = ioutil.WriteFile(localPath, body, 0644)
		if err != nil {
			log.Printf("Failed to write file: %v", err)
			continue
		}

		localPaths = append(localPaths, localPath)
	}

	return localPaths
}
func parseS3Link(link string) (bucket, key string, err error) {
	parts := strings.Split(link, "/")
	if len(parts) < 4 {
		return "", "", errors.New("formato de enlace S3 no válido")
	}

	bucket = parts[4] // El tercer elemento después de dividir por "/"
	key = parts[5]    // El resto de los elementos forman la clave (key)

	return bucket, key, nil
}
