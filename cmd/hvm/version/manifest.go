package version

import (
	_ "embed"
	"encoding/json"
)

//go:embed manifest.json
var manifestFile []byte
var manifest = &VersionManifest{}

type VersionManifest struct {
	Version string `json:"version"`
}

func GetVersionManifest() (*VersionManifest, error) {
	if *manifest == (VersionManifest{}) {
		err := json.Unmarshal(manifestFile, manifest)
		if err != nil {
			return nil, err
		}
	}

	return manifest, nil
}
