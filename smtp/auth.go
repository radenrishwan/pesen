package server

type SMTPAuth struct {
	Username string
	Password string
}

func NewSMTPAuth(username, password string) SMTPAuth {
	return SMTPAuth{
		Username: username,
		Password: password,
	}
}
