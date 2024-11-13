package main

import (
	"flag"
	"log"

	server "github.com/radenrishwan/smtp-from-scratch"
)

var (
	PORT = flag.String("port", "2525", "Port to run the SMTP server on. Default is 2525")
)

func main() {
	flag.Parse()

	s := server.NewServer(*PORT, true)

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
