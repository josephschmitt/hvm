package run

import (
	"github.com/josephschmitt/hvm"
	"github.com/josephschmitt/hvm/context"
)

type RunCmd struct {
	Name string   `kong:"arg"`
	Args []string `kong:"arg,optional"`

	Bin string
}

func (c *RunCmd) Run(ctx *context.Context) error {
	bin := c.Bin
	if bin == "" {
		bin = c.Name
	}

	return hvm.Run(ctx, c.Name, bin, c.Args...)
}
