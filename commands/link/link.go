package link

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/alecthomas/colour"
	"github.com/josephschmitt/hvm/manifest"
	"github.com/kardianos/osext"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasttemplate"
)

type LinkCmd struct {
	Name []string `kong:"arg,help='Project to link to your global bin'"`
}

func (c *LinkCmd) Run() error {
	man, _ := manifest.GetManifest()

	t := fasttemplate.New(string(runscriptFile), "{{", "}}")
	str := t.ExecuteString(map[string]interface{}{
		"version": man.Version,
	})

	for _, name := range c.Name {
		binPath, _ := osext.Executable()
		depPath := filepath.Join(filepath.Dir(binPath), name)

		err := createRunFile(depPath, str)
		if err != nil {
			return err
		}

		log.Infof(colour.Sprintf("Linked at: ^3%s^R", depPath))
	}

	return nil
}

//go:embed runscript.tmpl
var runscriptFile []byte

func createRunFile(path string, str string) error {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, []byte(str), 0755)
	if err != nil {
		return err
	}

	log.Debugf(colour.Sprintf("Write run file for ^3%s^R:\n%s\n", path, str))

	return nil
}
