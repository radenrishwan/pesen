package main

import (
	"fmt"
	"net/smtp"
)

const CRLF = "\r\n"

func main() {
	serverAddress := "localhost:2525"

	from := "sender@example.com"
	to := []string{"ujang@example.com", "agus@example.com"}
	subject := "subject gonna be here"
	body := "this is body of the email"

	message := []byte(
		"From: " + from + CRLF +
			"To: " + to[0] + CRLF +
			"Subject: " + subject + CRLF + CRLF +
			body + CRLF,
	)

	auth := smtp.PlainAuth("",
		"test",
		"test",
		"localhost",
	)

	err := smtp.SendMail(
		serverAddress,
		auth,
		// nil,
		from,
		to,
		message,
	)
	if err != nil {
		fmt.Println("Error sending email:", err)
		return
	}

	fmt.Println("Email sent successfully!")
}
