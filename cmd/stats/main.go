package main

import (
	"log"
	"net/http"

	"newsletter-app/internal/stats"
	"newsletter-app/pkg/db"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"os"
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

	statsHandler := stats.NewHandler(client)

	r := mux.NewRouter()
	r.Use(accessControlMiddleware)

	r.HandleFunc("/stats", statsHandler.GetStats).Methods("GET")

	log.Fatal(http.ListenAndServe(":8084", r))
}

// access control and  CORS middleware
func accessControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS,PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}
