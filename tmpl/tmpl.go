package tmpl

import (
	_ "embed"

	"github.com/valyala/fasttemplate"
)

const (
	TemplateMarker = "# HVM Script"
)

//go:embed runscript.tmpl
var runScriptFile []byte

func BuildRunScript(name string, bin string) string {
	t := fasttemplate.New(string(runScriptFile), "{{", "}}")
	return t.ExecuteString(map[string]interface{}{
		"marker": TemplateMarker,
		"name":   name,
		"bin":    bin,
	})
}
