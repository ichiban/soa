package gen

import (
	"bytes"
	_ "embed"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"slices"
	"strings"
	"text/template"

	"golang.org/x/tools/imports"
)

//go:embed soa.go.tmpl
var soaTemplate string

var fileTemplate = template.Must(template.New("").Funcs(map[string]any{
	"join": strings.Join,
}).Parse(soaTemplate))

type File struct {
	PackageName string
	Imports     []Import
	Structs     []Struct
}

func (f *File) WriteTo(w io.Writer) (int64, error) {
	var buf bytes.Buffer
	_ = fileTemplate.Execute(&buf, f)

	b, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		return 0, err
	}

	n, err := w.Write(b)
	return int64(n), err
}

type Import struct {
	Name string
	Path string
}

type Struct struct {
	Name      string
	SliceName string
	Fields    []Field
}

type Field struct {
	Names []string
	Type  string
}

func ParseFile(path string, target ...string) (File, error) {
	v := visitor{
		Target:  target,
		FileSet: token.NewFileSet(),
	}
	file, err := parser.ParseFile(v.FileSet, path, nil, parser.SkipObjectResolution)
	if err != nil {
		return File{}, err
	}

	ast.Walk(&v, file)
	return v.File, nil
}

type visitor struct {
	File

	Target  []string
	FileSet *token.FileSet
}

func (v *visitor) Visit(n ast.Node) ast.Visitor {
	switch n := n.(type) {
	case *ast.File:
		v.PackageName = n.Name.Name
		return v
	case *ast.ImportSpec:
		i := Import{Path: n.Path.Value}
		if n.Name != nil {
			i.Name = n.Name.Name
		}
		v.Imports = append(v.Imports, i)
		return nil
	case *ast.TypeSpec:
		if len(v.Target) > 0 && !slices.Contains(v.Target, n.Name.Name) {
			return v
		}

		t, ok := n.Type.(*ast.StructType)
		if !ok {
			return nil
		}

		fs := make([]Field, len(t.Fields.List))
		for i, f := range t.Fields.List {
			ns := make([]string, len(f.Names))
			for j, name := range f.Names {
				ns[j] = name.String()
			}
			var buf strings.Builder
			_ = printer.Fprint(&buf, v.FileSet, f.Type)
			fs[i] = Field{
				Names: ns,
				Type:  buf.String(),
			}
		}

		v.Structs = append(v.Structs, Struct{
			Name:   n.Name.Name,
			Fields: fs,
		})
		return nil
	default:
		return v
	}
}
