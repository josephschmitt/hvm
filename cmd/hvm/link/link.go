package link

import (
	_ "embed"

	"github.com/josephschmitt/hvm"
)

type LinkCmd struct {
	Name []string `kong:"arg,help='Project to link to your global bin'"`
}

func (c *LinkCmd) Run() error {
	return hvm.Link(c.Name)
}