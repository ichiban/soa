package main

import (
	_ "embed"
	"flag"
	"log"
	"os"

	"soa/internal/gen"
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

	if err := gen.Generate(in, out, name, flag.Args()...); err != nil {
		log.Fatal(err)
	}
}
