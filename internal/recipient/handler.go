package recipient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	client *mongo.Client
}

func NewHandler(client *mongo.Client) *Handler {
	return &Handler{client: client}
}

func (h *Handler) AddRecipient(w http.ResponseWriter, r *http.Request) {
	var recipient Recipient
	if err := json.NewDecoder(r.Body).Decode(&recipient); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	recipient.ID = uuid.New().String()
	collection := h.client.Database("emailDB").Collection("recipients")
	_, err := collection.InsertOne(context.TODO(), recipient)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Recipient added successfully")
}

func (h *Handler) UpdateRecipient(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var recipient Recipient
	if err := json.NewDecoder(r.Body).Decode(&recipient); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	collection := h.client.Database("emailDB").Collection("recipients")
	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.M{"$set": recipient})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Recipient updated successfully")
}

func (h *Handler) DeleteRecipient(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	collection := h.client.Database("emailDB").Collection("recipients")
	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Recipient deleted successfully")
}

func (h *Handler) GetAllRecipients(w http.ResponseWriter, r *http.Request) {
	var recipients []Recipient
	collection := h.client.Database("emailDB").Collection("recipients")
	cur, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var recipient Recipient
		err := cur.Decode(&recipient)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		recipients = append(recipients, recipient)
	}

	if err := cur.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(recipients)
}
