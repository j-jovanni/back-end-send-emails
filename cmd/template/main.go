package main

import (
	"log"
	"net/http"

	"newsletter-app/internal/template"
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

	templateHandler := template.NewHandler(client)

	r := mux.NewRouter()
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
	)(r)
	r.HandleFunc("/templates", templateHandler.CreateTemplate).Methods("POST")
	r.HandleFunc("/templates/{id}", templateHandler.GetTemplate).Methods("GET")
	r.HandleFunc("/templates", templateHandler.GetTemplates).Methods("GET")
	r.HandleFunc("/templates/{id}", templateHandler.UpdateTemplate).Methods("PUT")
	r.HandleFunc("/templates/{id}", templateHandler.DeleteTemplate).Methods("DELETE")
	r.HandleFunc("/templates/uploa_file", templateHandler.UploadHandler).Methods("POST")

	log.Fatal(http.ListenAndServe(":8083", corsHandler))
}
