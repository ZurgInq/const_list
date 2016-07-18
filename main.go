package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

var constList = `//CODE GENERATED AUTOMATICALLY
//THIS FILE SHOULD NOT BE EDITED BY HAND
package {{.Package}}

func {{.Name}}List() []{{.Name}} {
	list := []{{.Name}}{{"{"}}{{.List}}{{"}"}}

	return list
}
`

var (
	config struct {
		TypeName string
		Output   string
	}
)

func init() {
	flag.StringVar(&config.TypeName, "type", "", "type name")
	flag.StringVar(&config.Output, "output", "", "output file name")
	flag.Parse()
}

func main() {
	var source = flag.Arg(0)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, source, nil, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	packageDir := filepath.Dir(source)
	packageName := f.Name.Name

	consts := make([]string, 0)
	for _, decl := range f.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			switch decl.Tok {
			case token.CONST:
				for _, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.ValueSpec:
						if fmt.Sprint(spec.Type) == config.TypeName {
							consts = append(consts, spec.Names[0].Name)
						}
					}
				}
			}
		}
	}

	templateData := struct {
		Package string
		Name    string
		List    string
	}{
		Package: packageName,
		Name:    config.TypeName,
		List:    strings.Join(consts, ", "),
	}
	t := template.Must(template.New("const-list").Parse(constList))

	var outWriter io.Writer

	switch config.Output {
	case "stdout":
		outWriter = os.Stdout
	case "":
		outFilename := path.Join(packageDir, strings.ToLower(config.TypeName)+"_list.go")
		outWriter, err = os.Create(outFilename)
		if err != nil {
			panic(err)
		}
	default:
		outFilename := path.Join(packageDir, config.Output)
		outWriter, err = os.Create(outFilename)
		if err != nil {
			panic(err)
		}
	}

	if t.Execute(outWriter, templateData) != nil {
		fmt.Println(err)
	}
}
