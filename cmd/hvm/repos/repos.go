package repos

import (
	"github.com/josephschmitt/hvm"
	"github.com/josephschmitt/hvm/context"
)

type UpdateReposCmd struct{}

func (c *UpdateReposCmd) Run(ctx *context.Context) error {
	return hvm.UpdatePackagesRepos(ctx)
}
