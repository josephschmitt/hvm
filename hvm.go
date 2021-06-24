package hvm

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/colour"
	"github.com/josephschmitt/hvm/dep"
	"github.com/josephschmitt/hvm/paths"
	"github.com/josephschmitt/hvm/tmpl"
	"github.com/kardianos/osext"
	log "github.com/sirupsen/logrus"
)

var Paths *paths.Paths

func Link(names []string) error {
	script := tmpl.BuildRunScript()

	for _, name := range names {
		path := getScriptPath(name)

		err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
		if err != nil {
			return err
		}

		err = os.WriteFile(path, []byte(script), 0755)
		if err != nil {
			return err
		}

		log.Debugf(colour.Sprintf("Write run script:\n%s\n", script))
		log.Infof(colour.Sprintf("Linked to: ^3%s^R", path))
	}

	return nil
}

func UnLink(names []string, force bool) error {
	for _, name := range names {
		path := getScriptPath(name)
		file, err := os.Open(path)
		if os.IsNotExist(err) {
			log.Warn(colour.Sprintf("^2%s^R does not exist, skipping...", path))
			continue
		}

		if err != nil {
			return err
		}
		defer file.Close()

		isHVMManaged := isHVMScript(file)

		if isHVMManaged || force {
			if err := os.Remove(path); err != nil {
				return err
			}
		} else {
			errMsg := colour.Sprintf("attempting to unlink ^2%s^R which is NOT managed by HVM.\n", path) +
				colour.Sprintf("Use the --force flag if you wish to remove this file anyway.")
			return fmt.Errorf(errMsg)
		}

		if !isHVMManaged && force {
			log.Infof(colour.Sprintf("^1Forcibly^R un-linked from: ^3%s^R", path))
		} else {
			log.Infof(colour.Sprintf("Un-linked from: ^3%s^R", path))
		}
	}

	return nil
}

func Run(name string, args ...string) error {
	log.Debugf(colour.Sprintf("Run cmd [^2%s^R] with args: ^4%s^R\n",
		name, colour.Sprintf(strings.Join(args, "^R, ^4"))))

	conf := dep.GetDepConfig(name, Paths)
	dep, err := dep.ResolveDep(name, conf, Paths)
	if err != nil {
		return err
	}

	if err := Download(dep); err != nil {
		return err
	}

	return nil
}

func Download(dep *dep.Dependency) error {
	log.Debugf(colour.Sprintf("Download ^2%s^3\n", dep.Source))

	resp, err := http.Get(dep.Source)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dlFilePath := filepath.Join(Paths.TempDirectory, filepath.Base(dep.Source))

	// Create the file
	out, err := os.Create(dlFilePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	log.Debugf(colour.Sprintf("Downloaded file to ^6%s^R\n", dlFilePath))

	return nil
}

func getScriptPath(name string) string {
	binPath, _ := osext.Executable()
	return filepath.Join(filepath.Dir(binPath), name)
}

func isHVMScript(file io.Reader) bool {
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		txt := scanner.Text()
		if strings.HasPrefix(txt, tmpl.TemplateMarker) {
			return true
		}
	}

	return false
}

func init() {
	pths, err := paths.NewPaths()
	if err != nil {
		panic(err)
	}

	Paths = pths
}
