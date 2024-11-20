package server

import "strings"

type Mail struct {
	From   string
	To     []string
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

func (m *Mail) SetFrom(from string) *Mail {
	m.From = from

	return m
}

func (m *Mail) AddTo(to string) *Mail {
	m.To = append(m.To, to)

	return m
}
