package add

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

type AddCmd struct {
	Name []string `kong:"arg,help='Project to add'"`
}

func (c *AddCmd) Run() error {
	log.Printf("Add new dependencies: %s\n", strings.Join(c.Name, ", "))
	return nil
}
