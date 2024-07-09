package stats

import (
	"context"
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	client *mongo.Client
}

func NewHandler(client *mongo.Client) *Handler {
	return &Handler{client: client}
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	var stats struct {
		TotalEmailsSent   int `json:"total_emails_sent"`
		TotalRecipients   int `json:"total_recipients"`
		TotalUnsubscribed int `json:"total_unsubscribed"`
	}

	emailCollection := h.client.Database("emailDB").Collection("emails")
	recipientCollection := h.client.Database("emailDB").Collection("recipients")

	emailCount, err := emailCollection.CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	stats.TotalEmailsSent = int(emailCount)

	recipientCount, err := recipientCollection.CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	stats.TotalRecipients = int(recipientCount)

	unsubscribedCount, err := recipientCollection.CountDocuments(context.TODO(), bson.M{"unsubscribed": true})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	stats.TotalUnsubscribed = int(unsubscribedCount)

	json.NewEncoder(w).Encode(stats)
}
