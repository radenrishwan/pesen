package server

import "strings"

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
