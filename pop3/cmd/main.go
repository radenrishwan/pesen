package main

import (
	"flag"
	"log"

	"github.com/radenrishwan/pop3"
)

var auth = map[string]pop3.Auth{
	"raden": pop3.NewAuth("raden", "raden"),
	"test":  pop3.NewAuth("test", "test"),
}

var (
	// DEFAULT PORT FOR POP3 is 110
	PORT = flag.String("port", "1100", "Port to run the POP3 server on. Default is 1100.")
)

func main() {
	flag.Parse()

	server := pop3.NewServer(":" + *PORT)

	for _, v := range auth {
		server.AddAuth(v)
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}
