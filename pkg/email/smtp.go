package email

import (
	"crypto/tls"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

func init() {
	// Cargar variables desde el archivo .env
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func SendEmail(from string, to string, subject string, body string, attachments []string) error {

	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPort := 587
	smtpUser := os.Getenv("SMTP_USER")         // Your SMTP user
	smtpPassword := os.Getenv("SMTP_PASSWORD") // Your SMTP password

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	for _, file := range attachments {
		m.Attach(file)
	}

	d := gomail.NewDialer(
		smtpServer,
		smtpPort,
		smtpUser,
		smtpPassword,
	)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		log.Fatalf("Error sending email: %v", err)
		return err
	}

	log.Println("Email sent successfully.")
	return nil
}
