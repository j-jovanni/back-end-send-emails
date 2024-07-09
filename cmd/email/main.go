package main

import (
	"log"
	"net/http"
	"newsletter-app/internal/email"
	"newsletter-app/pkg/db"
	"os"

	"github.com/joho/godotenv"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Server struct {
	emailService *email.EmailService
}

func init() {
	// Cargar variables desde el archivo .env
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func main() {
	client := db.ConnectDB(os.Getenv("MONGO_ENV"))

	emailHandler := email.NewHandler(client)

	r := mux.NewRouter()

	//r.Use(accessControlMiddleware)
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
	)(r)

	r.HandleFunc("/send", emailHandler.SendEmail).Methods("POST")
	r.HandleFunc("/send-list", emailHandler.SendEmailToList).Methods("POST")
	r.HandleFunc("/unsubscribe/{templateID}/{email}", emailHandler.Unsubscribe).Methods("GET")
	// var emailService email.EmailService
	// emailService.ScheduleEmails()

	emailService := email.NewEmailService(client)
	server := &Server{
		emailService: emailService,
	}

	// Llamar a ScheduleEmails
	go server.emailService.ScheduleEmails()

	log.Fatal(http.ListenAndServe(":8081", corsHandler))
}
