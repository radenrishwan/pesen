package pop3

type SessionState struct {
	isAuthenticated bool
	username        string
	shouldQuit      bool
}

func NewSessionState() *SessionState {
	return &SessionState{
		isAuthenticated: false,
		username:        "",
		shouldQuit:      false,
	}
}
