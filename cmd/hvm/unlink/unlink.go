package unlink

import (
	"github.com/josephschmitt/hvm"
	"github.com/josephschmitt/hvm/context"
)

type UnLinkCmd struct {
	Name  []string `kong:"arg,help='Project(s) to unlink from your global bin.'"`
	Force bool     `kong:"help='Force unlink script(s), even if not managed by HVM.'"`
}

func (c *UnLinkCmd) Run(ctx *context.Context) error {
	return hvm.UnLink(ctx, c.Name, c.Force)
}
