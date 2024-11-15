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
	writer := bufio.NewWriter(conn)

	fmt.Fprintf(writer, "+OK POP3 server ready\r\n")

	for scanner.Scan() {
		line := scanner.Text()

		command := Command{}
		err := command.Parse(line)
		if err != nil {
			writer.WriteString(ERR + " " + err.Error() +
				"\r\n")
			continue
		}

		if command.Command != "" {
			log.Println("Client:", strings.TrimSpace(line))
		}

		if state.shouldQuit {
			log.Println("Closing connection")
			break
		}

		switch strings.ToUpper(command.Command) {
		case POP3_COMMAND_USER:
			if state.isAuthenticated {
				writer.WriteString(ERR + " Already authenticated\r\n")
				continue
			}

			if command.Args == "" {
				writer.WriteString(ERR + " Missing username\r\n")
				continue
			}

			state.username = command.Args

			writer.WriteString(OK + " User accepted\r\n")
		case POP3_COMMAND_PASS:
			if state.isAuthenticated {
				writer.WriteString(ERR + " Already authenticated\r\n")
				continue
			}

			if command.Args == "" {
				writer.WriteString(ERR + " Missing password\r\n")
				continue
			}

			if !validateAuth(state.username, command.Args) {
				writer.WriteString(ERR + " Invalid username or password\r\n")
				continue
			}
		case POP3_COMMAND_STAT:
			if !state.isAuthenticated {
				writer.WriteString(ERR + " Not authenticated\r\n")
				continue
			}

			// TODO: Implement later
			messageCount := 0
			size := 0

			writer.WriteString(OK + " " +
				strconv.Itoa(messageCount) + " " +
				strconv.Itoa(size) + "\r\n",
			)
		case POP3_COMMAND_LIST:
		case POP3_COMMAND_RETR:
		case POP3_COMMAND_NOOP:
		case POP3_COMMAND_DELE:
		case POP3_COMMAND_RSET:
		case POP3_COMMAND_QUIT:
		default:
			writer.WriteString(ERR + " Unknown command\r\n")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error reading from client:", err)
	}
}
