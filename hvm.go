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

	"github.com/josephschmitt/hvm/paths"
	"github.com/josephschmitt/hvm/repos"

	"github.com/alecthomas/colour"
	"github.com/josephschmitt/hvm/context"
	"github.com/josephschmitt/hvm/manifest"
	"github.com/josephschmitt/hvm/tmpl"
	log "github.com/sirupsen/logrus"
)

func Link(ctx *context.Context, names []string, force bool) error {
	loader := repos.NewGitRepoLoader("", "")
	loader.Update()
	exitCode := 0

	for _, name := range names {
		if !loader.HasPackage(name) {
			log.Errorf(colour.Sprintf("Package \"^3%s^R\" not found in package repository at ^6%s^R",
				name, loader.GetLocation()))
			exitCode++
			continue
		}

		var bins []string

		if manConf, err := manifest.NewPackageManfiestConfig(name); err == nil {
			for k := range manConf.Bins {
				bins = append(bins, k)
			}
		} else {
			bins = append(bins, name)
		}

		for _, bin := range bins {
			script := tmpl.BuildRunScript(name, bin)
			path := filepath.Join(ctx.LinkDir, bin)

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

	os.Exit(exitCode)
	return nil
}

func UnLink(ctx *context.Context, names []string, force bool) error {
	for _, name := range names {
		path := filepath.Join(ctx.LinkDir, name)
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
	manCtx := manifest.NewManifestContext(name, ctx.Use[bin])

	man, err := manifest.NewPackageManfiest(name, manCtx, ctx.Packages[name])
	if err != nil {
		return err
	}

	if !hasPackageLocally(manCtx.OutputDir, man.Bins[bin]) {
		if err := DownloadAndExtractPackage(ctx, man, manCtx); err != nil {
			return err
		}
	}

	cmdName := filepath.Join(manCtx.OutputDir, man.Bins[bin])
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

func DownloadAndExtractPackage(
	ctx *context.Context,
	man *manifest.PackageManifest,
	manCtx *manifest.PackageManifestContext,
) error {
	name := man.Name
	version := man.Version
	source := man.Source
	extract := man.Extract

	log.Infof(colour.Sprintf("Downloading ^3%s@%s^R from ^2%s^R...\n", name, version, source))

	if version == "" {
		return fmt.Errorf("no version set for package \"%s\", please set a version in config.hcl",
			name)
	}
	if source == "" || strings.Contains(source, "${") {
		return fmt.Errorf("no source URL set for package \"%s\"", name)
	}

	resp, err := http.Get(source)
	if err != nil {
		return err
	} else if resp.StatusCode >= 400 {
		return fmt.Errorf(colour.Sprintf("failed to download ^3%s@%s^R from ^1%s^R...", name,
			version, source))
	}
	defer resp.Body.Close()

	dlFilePath := filepath.Join(paths.AppPaths.TempDirectory, filepath.Base(source))
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

	outDir := manCtx.OutputDir

	err = os.MkdirAll(outDir, os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.Open(dlFilePath)
	if err != nil {
		return err
	}

	if extract != "" {
		extractCmdParts := strings.Split(extract, " ")
		extractCmd := extractCmdParts[0]
		extractArgs := extractCmdParts[1:]

		log.Debugf("Extract: %s", extract)

		cmd := exec.Command(extractCmd, extractArgs...)
		cmd.Dir = paths.AppPaths.TempDirectory
		cmd.Stdout = nil
		cmd.Stderr = os.Stderr
		cmd.Stdin = file

		if err := cmd.Run(); err != nil {
			return err
		}

		log.Debugf(colour.Sprintf("Successfully extracted to ^3%s^R\n", outDir))
	} else {
		outputPath := filepath.Join(outDir, name)
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

		log.Debugf(colour.Sprintf("No extract in manifest, moved download to ^3%s^R\n", outDir))
	}

	return nil
}

func hasPackageLocally(outdir string, bin string) bool {
	binPath := filepath.Join(outdir, bin)

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
