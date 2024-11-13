package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"strings"

	server "github.com/radenrishwan/smtp-from-scratch"
)

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

		fmt.Println("Connection accepted")

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

	reply(writer, server.SMTP_STATUS_READY, "Service ready")

	mail := NewMail()

	// print reader
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			slog.Error("Error reading from connection", "ERROR", err.Error())
			return
		}

		command := Command{}
		command.Parse(line)

		fmt.Println("Client:", strings.TrimSpace(line))

		if strings.HasPrefix(strings.ToUpper(command.Command), server.SMTP_COMMAND_HELO) {
			reply(writer, server.SMTP_STATUS_OK, "HELO from server")

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), server.SMTP_COMMAND_EHLO) {
			reply(writer, server.SMTP_STATUS_OK, "EHLO from server")

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), server.SMTP_COMMAND_MAIL) {
			reply(writer, server.SMTP_STATUS_OK, "MAIL command accepted")

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), server.SMTP_COMMAND_RCPT) {
			reply(writer, server.SMTP_STATUS_OK, "RCPT command accepted")

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), server.SMTP_COMMAND_DATA) {
			reply(writer, server.SMTP_STATUS_SEND_DATA, "End data with <CR><LF>.<CR><LF>")

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

			reply(writer, server.SMTP_STATUS_OK, "Mail accepted")

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), server.SMTP_COMMAND_RSET) {
			reply(writer, server.SMTP_STATUS_OK, "Resetting")

			mail = NewMail()

			// clear the reader
			reader.Reset(conn)

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), server.SMTP_COMMAND_NOOP) {
			reply(writer, server.SMTP_STATUS_OK, "Server is here...")

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), server.SMTP_COMMAND_QUIT) {
			reply(writer, server.SMTP_STATUS_BYE, "Dadah!")
			return
		}
	}
}

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

type Mail struct {
	Header map[string]string
	Body   string
}

func NewMail() Mail {
	return Mail{
		Header: make(map[string]string),
	}
}

func (m *Mail) Parse(data string) {
	lines := strings.Split(data, "\r\n")

	for i, line := range lines {
		if line == "" {
			m.Body = strings.Join(lines[i+1:], "\r\n")
			break
		}

		parts := strings.SplitN(line, ":", 2)
		m.Header[parts[0]] = strings.TrimSpace(parts[1])
	}
}

func reply(writer *bufio.Writer, code int, message string) {
	response := fmt.Sprintf("%d %s\r\n", code, message)

	writer.WriteString(response)
	writer.Flush()

	fmt.Println("Server:", strings.TrimSpace(response))
}
