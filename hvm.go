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

	"github.com/josephschmitt/hvm/repos"

	"github.com/alecthomas/colour"
	"github.com/josephschmitt/hvm/context"
	"github.com/josephschmitt/hvm/paths"
	"github.com/josephschmitt/hvm/pkgs"
	"github.com/josephschmitt/hvm/tmpl"
	"github.com/kardianos/osext"
	log "github.com/sirupsen/logrus"
)

func Link(ctx *context.Context, names []string, force bool) error {
	for _, name := range names {
		var bins []string

		man := &pkgs.PackageManifest{}

		if _, err := man.Resolve(name, ctx.Packages[name], paths.AppPaths); err == nil {
			for k := range man.Bins {
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

			file, err := os.Open(path)
			if err != nil && !os.IsNotExist(err) {
				return err
			}
			defer file.Close()

			isHVMManaged := os.IsNotExist(err) || isHVMScript(file)

			if isHVMManaged || force {
				if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
					return err
				}
			} else {
				errMsg := colour.Sprintf("attempting to link to an existing bin ^2%s^R which is NOT "+
					"managed by HVM.\n", path) +
					colour.Sprintf("Use the --overwrite flag if you wish to overwrite this file.")
				return fmt.Errorf(errMsg)
			}

			err = os.WriteFile(path, []byte(script), 0755)
			if err != nil {
				return err
			}

			log.Debugf(colour.Sprintf("Write run script:\n%s\n", script))

			if !isHVMManaged && force {
				log.Infof(colour.Sprintf("^1Forcibly^R overwrote: ^3%s^R", path))
			} else {
				log.Infof(colour.Sprintf("Linked to: ^3%s^R", path))
			}
		}
	}

	return nil
}

func UnLink(ctx *context.Context, names []string, force bool) error {
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

func Run(ctx *context.Context, name string, bin string, args ...string) error {
	man := &pkgs.PackageManifest{}
	if _, err := man.Resolve(name, ctx.Packages[name], paths.AppPaths); err != nil {
		return err
	}

	if !hasPackageLocally(man, bin) {
		if err := DownloadAndExtractPackage(ctx, man); err != nil {
			return err
		}
	}

	cmdName := filepath.Join(man.OutputDir, man.Bins[bin])
	if man.Exec != "" {
		args = append([]string{cmdName}, args...)
		cmdName = man.Exec
	}

	log.Debugf(colour.Sprintf("Run ^3%s^R@%s^R with args ^5%s^R\n", cmdName, man.Version, args))

	cmd := exec.Command(cmdName, args...)
	cmd.Dir = paths.AppPaths.WorkingDirectory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	log.Infof(colour.Sprintf("Using Hermetic ^3%s@%s^R\n", man.Name, man.Version))
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func GetPackageRepos(ctx *context.Context) error {
	loader := repos.NewGitRepoLoader("", "")
	return loader.Get()
}

func UpdatePackagesRepos(ctx *context.Context) error {
	loader := repos.NewGitRepoLoader("", "")
	return loader.Update()
}

func DownloadAndExtractPackage(ctx *context.Context, man *pkgs.PackageManifest) error {
	log.Infof(colour.Sprintf("Downloading ^3%s@%s^R from ^2%s^R...\n", man.Name, man.Version,
		man.Source))

	if man.Version == "" {
		return fmt.Errorf("invalid version \"%s\" for package \"%s\"", man.Version, man.Name)
	}
	if man.Source == "" || strings.Contains(man.Source, "${") {
		return fmt.Errorf("invalid source \"%s\" for package \"%s\"", man.Source, man.Name)
	}

	resp, err := http.Get(man.Source)
	if err != nil {
		return err
	} else if resp.StatusCode >= 400 {
		return fmt.Errorf(colour.Sprintf("failed to download ^3%s@%s^R from ^1%s^R...", man.Name,
			man.Version, man.Source))
	}
	defer resp.Body.Close()

	dlFilePath := filepath.Join(paths.AppPaths.TempDirectory, filepath.Base(man.Source))
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

	log.Debugf(colour.Sprintf("Downloaded file to ^6%s^R\n", dlFilePath))

	err = os.MkdirAll(man.OutputDir, os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.Open(dlFilePath)
	if err != nil {
		return err
	}

	if man.Extract != "" {
		extractCmdParts := strings.Split(man.Extract, " ")
		extractCmd := extractCmdParts[0]
		extractArgs := extractCmdParts[1:]

		log.Debugf("Extract: %s", man.Extract)

		cmd := exec.Command(extractCmd, extractArgs...)
		cmd.Dir = paths.AppPaths.TempDirectory
		cmd.Stdout = nil
		cmd.Stderr = os.Stderr
		cmd.Stdin = file

		if err := cmd.Run(); err != nil {
			return err
		}

		log.Debugf(colour.Sprintf("Successfully extracted to ^3%s^R\n", man.OutputDir))
	} else {
		outputPath := filepath.Join(man.OutputDir, man.Name)
		outputFile, err := os.Create(outputPath)
		if err != nil {
			return err
		}
		defer outputFile.Close()

		_, err = io.Copy(outputFile, file)
		if err != nil {
			return err
		}

		if err := os.Chmod(outputPath, 0755); err != nil {
			return err
		}

		log.Debugf(colour.Sprintf("No extract in manifest, moved download to ^3%s^R\n", man.OutputDir))
	}

	return nil
}

func getScriptPath(name string) string {
	binPath, _ := osext.Executable()
	return filepath.Join(filepath.Dir(binPath), name)
}

func hasPackageLocally(man *pkgs.PackageManifest, bin string) bool {
	binPath := filepath.Join(man.OutputDir, man.Bins[bin])

	if _, err := os.ReadFile(binPath); err == nil {
		return true
	}

	return false
}

func isHVMScript(file io.Reader) bool {
	if file == nil {
		return false
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		txt := scanner.Text()
		if strings.HasPrefix(txt, tmpl.TemplateMarker) {
			return true
		}
	}

	return false
}
