package run

import (
	"strings"

	"github.com/alecthomas/colour"
	log "github.com/sirupsen/logrus"
)

type RunCmd struct {
	Name string   `kong:"arg"`
	Args []string `kong:"arg,optional"`
}

func (c *RunCmd) Run() error {
	log.Printf(colour.Sprintf("Run cmd [^2%s^R] with args: ^4%s^R\n",
		c.Name, colour.Sprintf(strings.Join(c.Args, "^R, ^4"))))

	return nil
}
