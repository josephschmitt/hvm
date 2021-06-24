package run

import (
	"github.com/josephschmitt/hvm"
)

type RunCmd struct {
	Name string   `kong:"arg"`
	Args []string `kong:"arg,optional"`

	Bin string
}

func (c *RunCmd) Run() error {
	bin := c.Bin
	if bin == "" {
		bin = c.Name
	}

	return hvm.Run(c.Name, bin, c.Args...)
}
