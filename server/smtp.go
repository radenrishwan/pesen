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

	SMTP_STATUS_AUTH_SUCCESS = 235
)

const (
	SMTP_COMMAND_HELO = "HELO"
	SMTP_COMMAND_EHLO = "EHLO"
	SMTP_COMMAND_AUTH = "AUTH"
	SMTP_COMMAND_MAIL = "MAIL"
	SMTP_COMMAND_RCPT = "RCPT"
	SMTP_COMMAND_DATA = "DATA"
	SMTP_COMMAND_RSET = "RSET"
	SMTP_COMMAND_NOOP = "NOOP"
	SMTP_COMMAND_QUIT = "QUIT"
)

type Server struct {
	address   string
	auth      bool
	smtpAuths map[string]*SMTPAuth
}

func NewServer(address string, auth bool) *Server {
	// add dummy auth
	smtpAuths := make(map[string]*SMTPAuth)
	smtpAuths["test"] = NewSMTPAuth("test", "test")
	smtpAuths["test2"] = NewSMTPAuth("test2", "test2")

	return &Server{
		address:   address,
		auth:      auth,
		smtpAuths: smtpAuths,
	}
}

func (s *Server) ValidateAuth(username, password string) bool {
	auth, ok := s.smtpAuths[username]
	if !ok {
		return false
	}

	return auth.Password == password
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

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			slog.Error("Error reading from connection", "ERROR", err.Error())
			return
		}

		command := Command{}
		command.Parse(line)

		fmt.Println("Client:", strings.TrimSpace(line))

		if strings.HasPrefix(strings.ToUpper(command.Command), "*") {
			handleEhlo(writer, s)

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), SMTP_COMMAND_HELO) {
			handleHelo(writer, s)

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), SMTP_COMMAND_EHLO) {
			handleEhlo(writer, s)

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), SMTP_COMMAND_AUTH) {
			handleAuth(writer, s, command)

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), SMTP_COMMAND_MAIL) {
			handleMail(writer, &mail, command)

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), SMTP_COMMAND_RCPT) {
			handleRcpt(writer, &mail, command)

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), SMTP_COMMAND_DATA) {
			handleData(writer, reader, &mail)

			continue
		}

		if strings.HasPrefix(strings.ToUpper(command.Command), SMTP_COMMAND_RSET) {
			handleRset(writer)
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
			handleQuit(writer, &mail)

			return
		}
	}
}

func (s *Server) validateAuth(username string, password string) error {
	if s.auth {
		for _, auth := range s.smtpAuths {
			if auth.Username == username && auth.Password == password {
				return nil
			}
		}

	}

	return nil
}

func reply(writer *bufio.Writer, code int, message string) {
	response := fmt.Sprintf("%d %s\r\n", code, message)

	writer.WriteString(response)
	writer.Flush()

	fmt.Println("Server:", strings.TrimSpace(response))
}

func replyAuth(writer *bufio.Writer, code int, message string) {
	response := fmt.Sprintf("%d-%s\r\n", code, message)

	writer.WriteString(response)
	writer.Flush()

	fmt.Println("Server:", strings.TrimSpace(response))
}

func replyMultiLine(writer *bufio.Writer, code int, messages []string) {
	for i, msg := range messages {
		var response string
		if i == len(messages)-1 {
			response = fmt.Sprintf("%d %s\r\n", code, msg)
		} else {
			response = fmt.Sprintf("%d-%s\r\n", code, msg)
		}
		writer.WriteString(response)
	}
	writer.Flush()
}
