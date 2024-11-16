package pop3

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
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
