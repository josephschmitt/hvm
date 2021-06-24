package unlink

import (
	"github.com/josephschmitt/hvm"
)

type UnLinkCmd struct {
	Name  []string `kong:"arg,name='unlink',help='Project(s) to unlink from your global bin.'"`
	Force bool     `kong:"help='Force unlink script(s), even if not managed by HVM.'"`
}

func (c *UnLinkCmd) Run() error {
	return hvm.UnLink(c.Name, c.Force)
}
