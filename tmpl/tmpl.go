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
var runScriptFile []byte

func BuildRunScript(name string, bin string) string {
	man, _ := manifest.GetManifest()

	t := fasttemplate.New(string(runScriptFile), "{{", "}}")
	return t.ExecuteString(map[string]interface{}{
		"marker":  TemplateMarker,
		"version": man.Version,
		"name":    name,
		"bin":     bin,
	})
}
