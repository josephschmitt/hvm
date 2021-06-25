package context

import (
	"path/filepath"

	"github.com/josephschmitt/hvm/paths"
	log "github.com/sirupsen/logrus"
)

type Context struct {
	Debug log.Level
}

func ConfigDirs(pths *paths.Paths) []string {
	return []string{
		filepath.Join(pths.WorkingDirectory, ".hvm"),
		filepath.Join(pths.GitRoot, ".hvm"),
		filepath.Join(pths.HomeDirectory, ".hvm"),
	}
}

func ConfigFiles(pths *paths.Paths) []string {
	var files []string
	for _, dir := range ConfigDirs(pths) {
		files = append(files, filepath.Join(dir, "config.hcl"))
	}

	return files
}
