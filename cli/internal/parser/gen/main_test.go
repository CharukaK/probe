package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestTemplateIncludesChildren(t *testing.T) {
	var out bytes.Buffer
	data := tplData{
		Package: "ast",
		Nodes: []tplNode{{
			GoName:   "SourceFileNode",
			KindName: "SourceFileKind",
			Type:     "source_file",
			Children: &tplField{
				GoType:   "[]*RequestNode",
				Multiple: true,
				Ctor:     "NewRequestNode",
				Kinds:    []string{"request"},
			},
		}},
	}

	if err := outAstTmpl.Execute(&out, data); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Children []*RequestNode",
		`case "request":`,
		"v.Children = append(v.Children, &c)",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("generated output does not contain %q", want)
		}
	}
}
