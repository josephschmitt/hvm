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
	"github.com/josephschmitt/hvm/paths"
	"github.com/josephschmitt/hvm/pkgs"
	"github.com/josephschmitt/hvm/tmpl"
	"github.com/kardianos/osext"
	log "github.com/sirupsen/logrus"
)

func Link(names []string) error {
	for _, name := range names {
		var bins []string

		if conf, err := pkgs.NewPackageConfig(name, &pkgs.Package{}, paths.AppPaths); err == nil {
			for k := range conf.Links {
				bins = append(bins, k)
			}
		} else {
			bins = append(bins, name)
		}

		for _, bin := range bins {
			script := tmpl.BuildRunScript(name, bin)
			path := getScriptPath(bin)

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

func Run(name string, bin string, args ...string) error {
	log.Debugf(colour.Sprintf("Run cmd [^2%s^R] with args: ^4%s^R\n",
		name, colour.Sprintf(strings.Join(args, "^R, ^4"))))

	pkg := pkgs.NewPackage(name, paths.AppPaths)
	conf, err := pkgs.NewPackageConfig(name, pkg, paths.AppPaths)
	if err != nil {
		return err
	}

	if !HasPackageLocally(conf, bin) {
		if err := DownloadAndExtract(conf, pkg); err != nil {
			return err
		}
	}

	cmd := exec.Command(filepath.Join(conf.OutputDir, conf.Links[bin]), args...)
	cmd.Dir = paths.AppPaths.WorkingDirectory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func HasPackageLocally(conf *pkgs.PackageConfig, bin string) bool {
	binPath := filepath.Join(conf.OutputDir, conf.Links[bin])

	if _, err := os.ReadFile(binPath); err == nil {
		return true
	}

	return false
}

func DownloadAndExtract(conf *pkgs.PackageConfig, pkg *pkgs.Package) error {
	log.Infof(colour.Sprintf("Downloading ^2%s^R...\n", conf.Source))

	resp, err := http.Get(conf.Source)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dlFilePath := filepath.Join(paths.AppPaths.TempDirectory, filepath.Base(conf.Source))
	err = os.MkdirAll(filepath.Dir(dlFilePath), os.ModePerm)
	if err != nil {
		return err
	}

	dl, err := os.Create(dlFilePath)
	if err != nil {
		return err
	}
	defer dl.Close()

	_, err = io.Copy(dl, resp.Body)
	if err != nil {
		return err
	}

	log.Infof(colour.Sprintf("Downloaded file to ^6%s^R\n", dlFilePath))

	if len(conf.Extract) != 0 {
		err := os.MkdirAll(conf.OutputDir, os.ModePerm)
		if err != nil {
			return err
		}

		file, err := os.Open(dlFilePath)
		if err != nil {
			return err
		}

		log.Debugf("Extract cmd: %s %s", conf.Extract[0], strings.Join(conf.Extract[1:], " "))

		cmd := exec.Command(conf.Extract[0], conf.Extract[1:]...)
		cmd.Dir = paths.AppPaths.TempDirectory
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = file

		if err := cmd.Run(); err != nil {
			return err
		}

		log.Infof(colour.Sprintf("Extracted to ^3%s^R\n", conf.OutputDir))
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
