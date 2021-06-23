package manifest

import (
	_ "embed"
	"encoding/json"
)

//go:embed manifest.json
var manifestFile []byte
var manifest = &Manifest{}

type Manifest struct {
	Version string `json:"version"`
}

func GetManifest() (*Manifest, error) {
	if *manifest == (Manifest{}) {
		err := json.Unmarshal(manifestFile, manifest)
		if err != nil {
			return nil, err
		}
	}

	return manifest, nil
}
