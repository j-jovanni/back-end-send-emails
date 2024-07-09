package main

import (
	"log"
	"net/http"

	"newsletter-app/internal/recipient"
	"newsletter-app/pkg/db"

	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/joho/godotenv"
)

func init() {
	// Cargar variables desde el archivo .env
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func main() {
	client := db.ConnectDB(os.Getenv("MONGO_ENV"))

	recipientHandler := recipient.NewHandler(client)

	r := mux.NewRouter()
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
	)(r)

	r.HandleFunc("/recipients", recipientHandler.AddRecipient).Methods("POST")
	r.HandleFunc("/recipients/{id}", recipientHandler.UpdateRecipient).Methods("PUT")
	r.HandleFunc("/recipients/{id}", recipientHandler.DeleteRecipient).Methods("DELETE")
	r.HandleFunc("/recipients", recipientHandler.GetAllRecipients).Methods("GET")

	log.Fatal(http.ListenAndServe(":8082", corsHandler))
}
