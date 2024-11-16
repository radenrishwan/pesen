package pop3

import (
	"bufio"
	"fmt"
	"log"
	"log/slog"
	"net"
	"strconv"
	"strings"
)

const (
	OK  = "+OK"
	ERR = "-ERR"
)

var dummyMail = []*Mail{
	NewMail().
		SetFrom("raden@gmail.com").
		SetTo("agus@gmail.com").
		SetSubject("Sample mail 1").
		SetBody("This is a sample mail 1").
		AddHeader("Date", "2020-01-01").
		AddHeader("From", "raden@gmail.com").
		AddHeader("To", "agus@gmail.com"),
	NewMail().
		SetFrom("raden@gmail.com").
		SetTo("agus@gmail.com").
		SetSubject("Sample mail 2").
		SetBody("This is a sample mail 2").
		AddHeader("Date", "2020-01-02").
		AddHeader("From", "raden@gmail.com").
		AddHeader("To", "agus@gmail.com"),
	NewMail().
		SetFrom("raden@gmail.com").
		SetTo("acep@gmail.com").
		SetSubject("Sample mail 3").
		SetBody("This is a sample mail 3").
		AddHeader("Date", "2020-01-03").
		AddHeader("From", "raden@gmail.com").
		AddHeader("To", "acep@gmail.com"),
}

type Server struct {
	Addr string
	auth map[string]Auth
}

func NewServer(addr string) *Server {
	return &Server{
		Addr: addr,
		auth: make(map[string]Auth),
	}
}

func (s *Server) AddAuth(auth Auth) {
	if s.auth == nil {
		s.auth = make(map[string]Auth)
	}

	s.auth[auth.Username] = auth
}

func (s Server) GetAuth(username string) *Auth {
	a := s.auth[username]

	return &a
}

func (s Server) validateAuth(username, password string) bool {
	auth := s.GetAuth(username)
	if auth == nil {
		return false
	}

	return auth.Password == password
}

func (s *Server) ListenAndServe() error {
	if !strings.HasPrefix(s.Addr, ":") {
		s.Addr = ":" + s.Addr
	}

	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	defer listener.Close()

	slog.Info("Listening on " + s.Addr)

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

			if !s.validateAuth(state.username, command.Args) {
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
