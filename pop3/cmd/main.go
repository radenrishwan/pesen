package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/radenrishwan/pop3"
)

const (
	OK  = "+OK"
	ERR = "-ERR"
)

const (
	POP3_COMMAND_USER = "USER"
	POP3_COMMAND_PASS = "PASS"
	POP3_COMMAND_STAT = "STAT"
	POP3_COMMAND_LIST = "LIST"
	POP3_COMMAND_RETR = "RETR"
	POP3_COMMAND_NOOP = "NOOP"
	POP3_COMMAND_DELE = "DELE"
	POP3_COMMAND_RSET = "RSET"
	POP3_COMMAND_QUIT = "QUIT"
)

type Command struct {
	Command string
	Args    string
}

var dummyMail = []*pop3.Mail{
	pop3.NewMail().
		SetFrom("raden@gmail.com").
		SetTo("agus@gmail.com").
		SetSubject("Sample mail 1").
		SetBody("This is a sample mail 1").
		AddHeader("Date", "2020-01-01").
		AddHeader("From", "raden@gmail.com").
		AddHeader("To", "agus@gmail.com"),
	pop3.NewMail().
		SetFrom("raden@gmail.com").
		SetTo("agus@gmail.com").
		SetSubject("Sample mail 2").
		SetBody("This is a sample mail 2").
		AddHeader("Date", "2020-01-02").
		AddHeader("From", "raden@gmail.com").
		AddHeader("To", "agus@gmail.com"),
	pop3.NewMail().
		SetFrom("raden@gmail.com").
		SetTo("acep@gmail.com").
		SetSubject("Sample mail 3").
		SetBody("This is a sample mail 3").
		AddHeader("Date", "2020-01-03").
		AddHeader("From", "raden@gmail.com").
		AddHeader("To", "acep@gmail.com"),
}

func (c *Command) Parse(line string) error {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return errors.New("Empty command")
	}

	c.Command = parts[0]
	if len(parts) > 1 {
		c.Args = parts[1]
	}

	return nil
}

type SessionState struct {
	isAuthenticated bool
	username        string
	shouldQuit      bool
}

func NewSessionState() *SessionState {
	return &SessionState{
		isAuthenticated: false,
		username:        "",
		shouldQuit:      false,
	}
}

var auth = map[string]pop3.Auth{
	"raden": pop3.NewAuth("raden", "raden"),
	"test":  pop3.NewAuth("test", "test"),
}

var (
	// DEFAULT PORT FOR POP3 is 110
	PORT = flag.String("port", "1100", "Port to run the POP3 server on. Default is 1100.")
)

func validateAuth(username, password string) bool {
	_, ok := auth[username]
	if !ok {
		return false
	}

	return auth[username].Password == password
}

func main() {
	flag.Parse()

	server, err := net.Listen("tcp", ":"+*PORT)
	if err != nil {
		log.Panicln("Error when trying to listen on port", err)
	}

	defer server.Close()

	log.Println("Listening on port ", *PORT)

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	state := NewSessionState()
	scanner := bufio.NewScanner(conn)

	reply(conn, OK, "POP3 server ready")

	for scanner.Scan() {
		line := scanner.Text()

		command := Command{}
		err := command.Parse(line)
		if err != nil {
			reply(conn, ERR, err.Error())

			continue
		}

		log.Println("Client:", strings.TrimSpace(line))

		if state.shouldQuit {
			log.Println("Closing connection")
			break
		}

		switch strings.ToUpper(command.Command) {
		case POP3_COMMAND_USER:
			if state.isAuthenticated {
				reply(conn, ERR, "Already authenticated")

				continue
			}

			state.username = command.Args

			reply(conn, OK, "User accepted")
		case POP3_COMMAND_PASS:
			if state.isAuthenticated {
				reply(conn, OK, "Already authenticated")

				continue
			}

			if command.Args == "" {
				reply(conn, ERR, "Missing password")

				continue
			}

			if !validateAuth(state.username, command.Args) {
				reply(conn, ERR, "Invalid username or password")

				continue
			}

			state.isAuthenticated = true

			reply(conn, OK, "Authenticated")
		case POP3_COMMAND_STAT:
			if !state.isAuthenticated {
				reply(conn, ERR, "Not authenticated")

				continue
			}

			messageCount := len(dummyMail)
			size := 0
			for _, mail := range dummyMail {
				size += mail.Size()
			}

			reply(conn, OK, fmt.Sprintf("%d %d", messageCount, size))

			continue
		case POP3_COMMAND_LIST:
			if !state.isAuthenticated {
				reply(conn, ERR, "Not authenticated")

				continue
			}

			reply(conn, OK, fmt.Sprintf(""))

			// // check if has a argument
			// if command.Args != "" {
			// 	// get specific mail
			// 	index, err := strconv.Atoi(command.Args)
			// 	if err != nil {
			// 		reply(conn, ERR, "Invalid message number")
			// 	}

			// 	if index < 1 || index > len(dummyMail) {
			// 		reply(conn, ERR, "No such message")
			// 	}

			// 	replyWithoutStatus(conn, fmt.Sprintf("%d %d", index, dummyMail[index-1].Size()))

			// 	continue
			// }

			var messages []string
			for i, mail := range dummyMail {
				messages = append(messages, fmt.Sprintf("%d %d", i+1, mail.Size()))
			}

			replyMultiline(conn, messages, true)

			continue
		case POP3_COMMAND_RETR:
			if !state.isAuthenticated {
				reply(conn, ERR, "Not authenticated")

				continue
			}

			index, err := strconv.Atoi(command.Args)
			if err != nil {
				reply(conn, ERR, "Invalid message number")
			}

			if index < 1 || index > len(dummyMail) {
				reply(conn, ERR, "No such message")
			}

			mail := dummyMail[index-1]

			reply(conn, OK, strconv.Itoa(mail.Size()))

			replyWithoutStatus(conn, mail.String())
			replyWithoutStatus(conn, ".")
		case POP3_COMMAND_NOOP:
			reply(conn, OK, "NOOP")

			continue
		case POP3_COMMAND_DELE:
			if !state.isAuthenticated {
				reply(conn, ERR, "Not authenticated")

				continue
			}

			index, err := strconv.Atoi(command.Args)
			if err != nil {
				reply(conn, ERR, "Invalid message number")

				continue
			}

			if index < 1 || index > len(dummyMail) {
				reply(conn, ERR, "No such message")

				continue
			}

			dummyMail = append(dummyMail[:index-1], dummyMail[index:]...)

			reply(conn, OK, "Message deleted")
		case POP3_COMMAND_RSET:
			if !state.isAuthenticated {
				reply(conn, ERR, "Not authenticated")
			}

			// clear all buffer and state
			state = NewSessionState()
		case POP3_COMMAND_QUIT:
			reply(conn, OK, "Bye")

			break
		default:
			reply(conn, ERR, "Unknown command")

			continue
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error reading from client:", err)
	}
}

func reply(conn net.Conn, status string, message string) {
	fmt.Fprintf(conn, "%s %s\r\n", status, message)

	log.Printf("Server: %s %s\r\n", status, message)
}

func replyWithoutStatus(conn net.Conn, message string) {
	fmt.Fprintf(conn, "%s\r\n", message)

	log.Printf("Server: %s", message)
}

func replyMultiline(conn net.Conn, messages []string, end bool) {
	for _, message := range messages {
		replyWithoutStatus(conn, fmt.Sprintf("%s", message))
	}

	if end {
		replyWithoutStatus(conn, ".")
	}
}
