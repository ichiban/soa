package main

import (
	_ "embed"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ichiban/soa/internal/gen"
)

func main() {
	var (
		in   string
		out  string
		name string
	)
	flag.StringVar(&in, "in", os.Getenv("GOFILE"), "path to input file (default $GOFILE)")
	flag.StringVar(&out, "out", "{{dir .}}/{{stem .}}_soa{{ext .}}", "path to output file or - for stdout")
	flag.StringVar(&name, "name", "{{.}}Slice", "name of generated soa slice")
	flag.Parse()

	if err := Generate(in, out, name, flag.Args()...); err != nil {
		log.Fatal(err)
	}
}

func Generate(in, out, name string, target ...string) error {
	f, err := gen.ParseFile(in, target...)
	if err != nil {
		return err
	}

	n, err := template.New("").Parse(name)
	if err != nil {
		return err
	}

	var sb strings.Builder
	for i := range f.Structs {
		s := &f.Structs[i]

		sb.Reset()
		if err := n.Execute(&sb, s.Name); err != nil {
			return err
		}

		s.SliceName = sb.String()
	}

	var w io.Writer
	if out == "-" {
		w = os.Stdout
	} else {
		o, err := template.New("").Funcs(map[string]any{
			"dir": filepath.Dir,
			"stem": func(in string) string {
				base := filepath.Base(in)
				return strings.TrimSuffix(base, filepath.Ext(base))
			},
			"ext": filepath.Ext,
		}).Parse(out)
		if err != nil {
			return err
		}
		var sb strings.Builder
		if err := o.Execute(&sb, in); err != nil {
			return err
		}
		out = filepath.Clean(sb.String())

		f, err := os.Create(out)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}

	_, err = f.WriteTo(w)
	return err
}
