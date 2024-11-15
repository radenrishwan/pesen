package pop3

type Auth struct {
	Username string
	Password string
}

func NewAuth(username, password string) Auth {
	return Auth{
		Username: username,
		Password: password,
	}
}
