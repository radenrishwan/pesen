package main

import (
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
}
