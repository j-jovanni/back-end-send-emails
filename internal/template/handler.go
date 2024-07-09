package template

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

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

func (h *Handler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var template Template

	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	template.ID = uuid.New().String()
	template.FromEmail = "jesus.flores.aws@gmail.com"

	collection := h.client.Database("emailDB").Collection("templates")
	_, err := collection.InsertOne(context.TODO(), template)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Template created successfully")
}

func (h *Handler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var template Template
	collection := h.client.Database("emailDB").Collection("templates")
	err = collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&template)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(template)
}

func (h *Handler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	var templates []Template
	collection := h.client.Database("emailDB").Collection("templates")
	cur, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var template Template
		err := cur.Decode(&template)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		templates = append(templates, template)
	}

	if err := cur.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(templates)
}

func (h *Handler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var template Template
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	collection := h.client.Database("emailDB").Collection("templates")
	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.M{"$set": template})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Template updated successfully")
}

func (h *Handler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	collection := h.client.Database("emailDB").Collection("templates")
	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Template deleted successfully")
}
func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("could not get uploaded file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	key := header.Filename

	uploader := s3.PutObjectInput{
		Bucket:      aws.String("files-emails-flocar"),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(http.DetectContentType([]byte(ext))),
	}

	awsAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-2"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsAccessKeyID, awsSecretAccessKey, "")),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to load SDK config: %v", err), http.StatusBadRequest)

		log.Fatalf("unable to load SDK config, %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	_, err = s3Client.PutObject(context.TODO(), &uploader)
	if err != nil {
		http.Error(w, fmt.Sprintf("uFailed to upload file from S3: %v", err), http.StatusBadRequest)

		log.Printf("Failed to download file from S3: %v", err)

	}
	// _, err = s3Client.PutObject(context.TODO(), &uploader)
	// if err != nil {
	// 	http.Error(w, fmt.Sprintf("could not upload file to S3: %v", err), http.StatusInternalServerError)
	// 	return
	// }

	url := fmt.Sprintf("https://s3://%s/%s", "files-emails-flocar", key)
	fmt.Fprintf(w, url)
}
