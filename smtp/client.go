package server

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Auth interface {
	Validate() bool
}

type PlainAuth struct {
	Username string
	Password string
}

type Dialer struct {
	Host   string
	Port   string
	reader *bufio.Reader
	writer *bufio.Writer
}

func NewDialer(host, port string) *Dialer {
	return &Dialer{
		Host: host,
		Port: port,
	}
}

func (d *Dialer) Dial() (net.Conn, error) {
	addr := d.Host + ":" + d.Port
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (d *Dialer) Close(conn net.Conn) {
	conn.Close()
}

// use nil if the server does not require authentication
func (d *Dialer) SendMail(mail Mail, auth Auth) error {
	conn, err := d.Dial()
	if err != nil {
		return err
	}

	defer d.Close(conn)

	d.reader = bufio.NewReader(conn)
	d.writer = bufio.NewWriter(conn)

	// waiting for server to send 220
	line, err := d.reader.ReadString('\n')
	if err != nil {
		return err
	}

	// if the response is not 220, return error
	if !strings.HasPrefix(line, strconv.Itoa(SMTP_STATUS_READY)) {
		return errors.New("server did not respond with 220")
	}

	for {
		// waiting for server to send 220
		line, err := d.reader.ReadString('\n')
		if err != nil {
			return err
		}

		// if the response is not 220, return error
		if !strings.HasPrefix(line, strconv.Itoa(SMTP_STATUS_READY)) {
			return errors.New("server did not respond with 220")
		}

		// send ehlo
		d.reply(SMTP_COMMAND_EHLO, []string{d.Host})

		// waiting for server to send 250
		line, err = d.reader.ReadString('\n')
		if err != nil {
			return err
		}

		// if the response is not 250, return error
		if !strings.HasPrefix(line, strconv.Itoa(SMTP_STATUS_OK)) {
			return errors.New("server did not respond with 250")
		}

		// check if server needs authentication

		// send mail from

		// send rcpt to

		// send data

		// send mail body

		// send .

		// send quit
	}
}

func (d *Dialer) reply(command string, args []string) {
	response := command + " " + strings.Join(args, " ")

	d.writer.WriteString(response)
	d.writer.Flush()

	fmt.Println("Client: ", response)
}

func (d *Dialer) replyAuth(username, password string) {
	response := username + "\x00" + username + "\x00" + password

	d.writer.WriteString(response)
	d.writer.Flush()

	fmt.Println("Client: ", response)
}
