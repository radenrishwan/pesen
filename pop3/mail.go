package pop3

import "fmt"

type Mail struct {
	From    string
	To      string
	Header  map[string]string
	Subject string
	Body    string
}

func NewMail() *Mail {
	return &Mail{
		Header: make(map[string]string),
	}
}

func (m *Mail) AddHeader(key, value string) *Mail {
	m.Header[key] = value

	return m
}

func (m *Mail) SetBody(body string) *Mail {
	m.Body = body
	return m
}

func (m *Mail) SetSubject(subject string) *Mail {
	m.Subject = subject
	return m
}

func (m *Mail) SetFrom(from string) *Mail {
	m.From = from
	return m
}

func (m *Mail) SetTo(to string) *Mail {
	m.To = to
	return m
}

func (m *Mail) GetHeader(key string) string {
	return m.Header[key]
}

// get the byte size of the mail
func (m Mail) Size() int {
	size := 0

	if m.To != "" {
		size += len([]byte(m.To))
	}

	if m.From != "" {
		size += len([]byte(m.From))
	}

	if m.Subject != "" {
		size += len([]byte(m.Subject))
	}

	if m.Body != "" {
		size += len([]byte(m.Body))
	}

	for k, v := range m.Header {
		size += len([]byte(k))
		size += len([]byte(v))
	}

	return size
}

func (m Mail) String() string {
	result := ""

	// check if header has From and To
	if ok := m.Header["From"]; ok != "" {
		result += fmt.Sprintf("From: %s\r\n", m.From)
	}

	if ok := m.Header["To"]; ok != "" {
		result += fmt.Sprintf("To: %s\r\n", m.To)
	}

	if ok := m.Header["Subject"]; ok != "" {
		result += fmt.Sprintf("Subject: %s\r\n", m.Subject)
	}

	for k, v := range m.Header {
		result += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	// add body
	result += fmt.Sprintf("\r\n%s\r\n", m.Body)

	return result
}
