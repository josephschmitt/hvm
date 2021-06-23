package hvm

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/colour"
	"github.com/josephschmitt/hvm/tmpl"
	"github.com/kardianos/osext"
	log "github.com/sirupsen/logrus"
)

func Link(names []string) error {
	script := tmpl.BuildRunScript()

	for _, name := range names {
		binPath, _ := osext.Executable()
		depPath := filepath.Join(filepath.Dir(binPath), name)

		err := os.MkdirAll(filepath.Dir(depPath), os.ModePerm)
		if err != nil {
			return err
		}

		err = os.WriteFile(depPath, []byte(script), 0755)
		if err != nil {
			return err
		}

		log.Debugf(colour.Sprintf("Write run script:\n%s\n", script))
		log.Infof(colour.Sprintf("Linked to: ^3%s^R", depPath))
	}

	return nil
}

func Run(name string, args ...string) error {
	log.Printf(colour.Sprintf("Run cmd [^2%s^R] with args: ^4%s^R\n",
		name, colour.Sprintf(strings.Join(args, "^R, ^4"))))

	return nil
}
