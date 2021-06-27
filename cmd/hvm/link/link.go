package link

import (
	"github.com/josephschmitt/hvm"
	"github.com/josephschmitt/hvm/context"
)

type LinkCmd struct {
	Name []string `kong:"arg,help='Project(s) to link to your global bin.'"`
}

func (c *LinkCmd) Run(ctx *context.Context) error {
	return hvm.Link(ctx, c.Name)
}
