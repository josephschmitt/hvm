package tmpl

import (
	_ "embed"

	"github.com/josephschmitt/hvm/manifest"
	"github.com/valyala/fasttemplate"
)

const (
	TemplateMarker = "# HVM Script"
)

//go:embed runscript.tmpl
var runscriptFile []byte

func BuildRunScript() string {
	man, _ := manifest.GetManifest()

	t := fasttemplate.New(string(runscriptFile), "{{", "}}")
	return t.ExecuteString(map[string]interface{}{
		"marker":  TemplateMarker,
		"version": man.Version,
	})
}
