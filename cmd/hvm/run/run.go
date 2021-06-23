package run

import (
	"github.com/josephschmitt/hvm"
)

type RunCmd struct {
	Name string   `kong:"arg"`
	Args []string `kong:"arg,optional"`
}

func (c *RunCmd) Run() error {
	return hvm.Run(c.Name, c.Args...)
}
