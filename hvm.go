package hvm

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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

	if err := DownloadAndExtract(dep, conf); err != nil {
		return err
	}

	return nil
}

func DownloadAndExtract(dep *dep.Dependency, conf *dep.Config) error {
	log.Debugf(colour.Sprintf("Downloading ^2%s^R...\n", dep.Source))

	resp, err := http.Get(dep.Source)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dlFilePath := filepath.Join(Paths.TempDirectory, filepath.Base(dep.Source))

	dl, err := os.Create(dlFilePath)
	if err != nil {
		return err
	}
	defer dl.Close()

	_, err = io.Copy(dl, resp.Body)
	if err != nil {
		return err
	}

	log.Debugf(colour.Sprintf("Downloaded file to ^6%s^R\n", dlFilePath))

	if len(dep.Extract) != 0 {
		err := os.MkdirAll(conf.OutputDir, os.ModePerm)
		if err != nil {
			return err
		}

		file, err := os.Open(dlFilePath)
		if err != nil {
			return err
		}

		cmd := exec.Command(dep.Extract[0], dep.Extract[1:]...)
		cmd.Dir = Paths.TempDirectory
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = file

		if err := cmd.Run(); err != nil {
			return err
		}

		log.Debugf(colour.Sprintf("Extracted to ^3%s^R\n", conf.OutputDir))
	}

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
