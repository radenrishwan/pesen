package server

import (
	"bufio"
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

type Server struct {
	address string
}

func NewServer(address string) *Server {
	return &Server{
		address: address,
	}
}

func (s *Server) ListenAndServe() error {
	if !strings.HasPrefix(s.address, ":") {
		s.address = ":" + s.address
	}

	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("Error accepting connection", "ERROR", err.Error())
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	reply(writer, SMTP_STATUS_READY, "Service ready")

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

		if strings.HasPrefix(strings.ToUpper(command.Command), SMTP_COMMAND_HELO) {
			reply(writer, SMTP_STATUS_OK, "HELO from server")

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), SMTP_COMMAND_EHLO) {
			reply(writer, SMTP_STATUS_OK, "EHLO from server")

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), SMTP_COMMAND_MAIL) {
			reply(writer, SMTP_STATUS_OK, "MAIL command accepted")

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), SMTP_COMMAND_RCPT) {
			reply(writer, SMTP_STATUS_OK, "RCPT command accepted")

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), SMTP_COMMAND_DATA) {
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

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), SMTP_COMMAND_RSET) {
			reply(writer, SMTP_STATUS_OK, "Resetting")

			mail = NewMail()

			// clear the reader
			reader.Reset(conn)

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), SMTP_COMMAND_NOOP) {
			reply(writer, SMTP_STATUS_OK, "Server is here...")

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), SMTP_COMMAND_QUIT) {
			reply(writer, SMTP_STATUS_BYE, "Dadah!")
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
