package server

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"
)

type Command struct {
	Command string
	Args    []string
}

func (c *Command) Parse(line string) {
	parts := strings.Fields(line)

	// check if parts is empty
	if len(parts) == 0 {
		return
	}

	c.Command = strings.ToUpper(parts[0])

	c.Args = parts[1:]
}

func handleHelo(writer *bufio.Writer, s *Server) {
	if s.auth {
		replyMultiLine(writer, SMTP_STATUS_OK, []string{
			fmt.Sprintf("%s at your service, [127.0.0.1]", s.address),
			"AUTH PLAIN",
		})
	} else {
		reply(writer, SMTP_STATUS_OK, "HELO from server")
	}
}

func handleEhlo(writer *bufio.Writer, s *Server) {
	if s.auth {
		replyMultiLine(writer, SMTP_STATUS_OK, []string{
			fmt.Sprintf("%s at your service, [127.0.0.1]", s.address),
			"AUTH PLAIN",
		})
	} else {
		reply(writer, SMTP_STATUS_OK, "EHLO from server")
	}
}

func handleAuth(writer *bufio.Writer, s *Server, command Command) {
	if !s.auth {
		reply(writer, SMTP_STATUS_ERROR_SYNTAX, "Authentication not enabled")

		return
	}

	if len(command.Args) == 0 {
		reply(writer, SMTP_STATUS_ERROR_SYNTAX, "AUTH command requires an argument")

		return
	}

	fmt.Println("Command Args:", command.Args)

	// check if auth is not plain
	if strings.ToUpper(command.Args[0]) != "PLAIN" {
		reply(writer, SMTP_STATUS_ERROR_SYNTAX, "Only PLAIN authentication is supported")

		return
	}

	// decode base64
	decoded, err := base64.StdEncoding.DecodeString(command.Args[1])
	if err != nil {
		reply(writer, SMTP_STATUS_ERROR_SYNTAX, "Invalid base64 encoding")
	}

	// split username and password
	if valid := s.ValidateAuth(strings.Split(string(decoded), "\x00")[1], strings.Split(string(decoded), "\x00")[2]); !valid {
		reply(writer, SMTP_STATUS_ERROR_SYNTAX, "Invalid username or password")

		return
	}

	reply(writer, SMTP_STATUS_AUTH_SUCCESS, "Authentication successful")
}

func handleMail(writer *bufio.Writer, mail *Mail, command Command) {
	if len(command.Args) == 0 {
		reply(writer, SMTP_STATUS_ERROR_SYNTAX, "MAIL command requires an argument")

		return
	}

	s := strings.Split(command.Args[0], ":")
	r := strings.NewReplacer("<", "", ">", "")

	mail.SetFrom(r.Replace(s[1]))

	reply(writer, SMTP_STATUS_OK, "MAIL command accepted")
}

func handleRcpt(writer *bufio.Writer, mail *Mail, command Command) {
	if len(command.Args) == 0 {
		reply(writer, SMTP_STATUS_ERROR_SYNTAX, "RCPT command requires an argument")

		return
	}

	s := strings.Split(command.Args[0], ":")
	r := strings.NewReplacer("<", "", ">", "")

	mail.AddTo(r.Replace(s[1]))

	reply(writer, SMTP_STATUS_OK, "RCPT command accepted")
}

func handleData(writer *bufio.Writer, reader *bufio.Reader, mail *Mail) {
	reply(writer, SMTP_STATUS_SEND_DATA, "End data with <CR><LF>.<CR><LF>")

	data := ""

	for {
		line, err := reader.ReadString('\n')

		fmt.Println("Client:", strings.TrimSpace(line))

		if err != nil {
			slog.Error("Error reading from connection", "ERROR", err.Error())
			return
		}

		if strings.TrimSpace(line) == "." {
			break
		}

		data += line
	}

	mail.Parse(data)

	reply(writer, SMTP_STATUS_OK, "Mail accepted")
}

func handleRset(writer *bufio.Writer) {
	reply(writer, SMTP_STATUS_OK, "Resetting")
}

func handleNoop(writer *bufio.Writer) {
	reply(writer, SMTP_STATUS_OK, "I'm with you <3")
}

func handleQuit(writer *bufio.Writer, mail *Mail) {
	fmt.Println(mail)

	reply(writer, SMTP_STATUS_BYE, "Dadah!")
}
