package main

import (
	"io"
	"log"

	"github.com/knadh/go-pop3"
)

func main() {
	p := pop3.New(pop3.Opt{
		Host:       "localhost",
		Port:       1100,
		TLSEnabled: false,
	})

	conn, err := p.NewConn()
	if err != nil {
		log.Fatalln(err)
	}

	defer conn.Quit()

	// Authenticate
	if err := conn.Auth("raden", "raden"); err != nil {
		log.Fatal(err)
	}

	log.Println("Authenticated")

	// Stat
	n, size, err := conn.Stat()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("You have %d messages with a total size of %d bytes\n", n, size)

	// List
	list, err := conn.List(0)
	if err != nil {
		log.Fatal(err)
	}

	for _, l := range list {
		log.Printf("Message %d has a size of %d bytes\n", l.ID, l.Size)
	}

	// Retrieve
	msg, err := conn.Retr(1)
	if err != nil {
		log.Fatal(err)
	}

	b, _ := io.ReadAll(msg.Body)
	log.Println("Message body: ", string(b))
	log.Println("Message Header: ", msg.Header.Map())
}
