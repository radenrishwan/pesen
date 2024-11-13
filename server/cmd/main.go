package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"strings"
)

const (
	SMTP_STATUS_READY                      = 220
	SMTP_STATUS_BYE                        = 221
	SMTP_STATUS_OK                         = 250
	SMTP_STATUS_SEND_DATA                  = 354
	SMTP_STATUS_ERROR_COMMAND_UNRECOGNIZED = 500
	SMTP_STATUS_ERROR_SYNTAX               = 501
	SMTP_STATUS_ERROR_BAD_SEQUENCE         = 503
)

const (
	SMTP_COMMAND_HELO = "HELO"
	SMTP_COMMAND_EHLO = "EHLO"
	SMTP_COMMAND_MAIL = "MAIL"
	SMTP_COMMAND_RCPT = "RCPT"
	SMTP_COMMAND_DATA = "DATA"
	SMTP_COMMAND_RSET = "RSET"
	SMTP_COMMAND_NOOP = "NOOP"
	SMTP_COMMAND_QUIT = "QUIT"
)

type SMTPSession struct {
	sender     string
	recipients []string
	dataBuffer strings.Builder
	state      string
}

func (self *SMTPSession) Reset() {
	self.sender = ""
	self.recipients = nil
	self.dataBuffer.Reset()
	self.state = ""
}

var (
	PORT = flag.String("port", "2525", "Port to run the SMTP server on. Default is 2525")
)

func main() {
	flag.Parse()

	ln, err := net.Listen("tcp", ":"+*PORT)
	if err != nil {
		slog.Error("Error while trying to listen "+*PORT, "ERROR", err.Error())
		return
	}
	defer ln.Close()

	slog.Info("SMTP server is running", "PORT", *PORT)

	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Error("Error accepting connection", "ERROR", err.Error())
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	reply(writer, SMTP_STATUS_READY, "SMTP Ready")

	session := SMTPSession{
		state: "",
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			slog.Error("Error reading from connection", "ERROR", err.Error())
			return
		}

		// print the client message
		line = strings.TrimSpace(line)
		fmt.Println("Client:", line)

		switch session.state {
		case SMTP_COMMAND_DATA:
			if line == "." {
				fmt.Println("Email Data Received:")
				fmt.Println("From:", session.sender)
				fmt.Println("To:", session.recipients)
				fmt.Println(session.dataBuffer.String())

				session.Reset()

				reply(writer, SMTP_STATUS_OK, "OK: Message accepted for delivery")
			} else {
				session.dataBuffer.WriteString(line + "\r\n")
			}
		default:
			if strings.HasPrefix(strings.ToUpper(line), SMTP_COMMAND_HELO) {
				reply(writer, SMTP_STATUS_OK, "Hello")
			} else if strings.HasPrefix(strings.ToUpper(line), SMTP_COMMAND_EHLO) {
				reply(writer, SMTP_STATUS_OK, "Hello")
			} else if strings.HasPrefix(strings.ToUpper(line), SMTP_COMMAND_MAIL) {
				session.sender = parseAddress(line[10:])

				fmt.Println("MAIL FROM:", session.sender)

				reply(writer, SMTP_STATUS_OK, "OK")
			} else if strings.HasPrefix(strings.ToUpper(line), SMTP_COMMAND_RCPT) {
				recipient := parseAddress(line[8:])
				session.recipients = append(session.recipients, recipient)

				fmt.Println("RCPT TO:", recipient)

				reply(writer, SMTP_STATUS_OK, "OK")
			} else if strings.ToUpper(line) == SMTP_COMMAND_DATA {
				if session.sender == "" || len(session.recipients) == 0 {
					reply(writer, SMTP_STATUS_ERROR_BAD_SEQUENCE, "Bad sequence of commands")
				} else {
					reply(writer, SMTP_STATUS_SEND_DATA, "End data with <CR><LF>.<CR><LF>")

					session.state = "DATA"
				}
			} else if strings.ToUpper(line) == SMTP_COMMAND_QUIT {
				reply(writer, SMTP_STATUS_BYE, "Bye")

				return
			} else {
				reply(writer, SMTP_STATUS_ERROR_COMMAND_UNRECOGNIZED, "Syntax error, command unrecognized")
			}
		}
	}
}

func reply(writer *bufio.Writer, code int, message string) {
	response := fmt.Sprintf("%d %s\r\n", code, message)
	writer.WriteString(response)
	writer.Flush()
	fmt.Println("Server:", strings.TrimSpace(response))
}

func parseAddress(address string) string {
	address = strings.TrimSpace(address)
	address = strings.Trim(address, "<>")
	return address
}
