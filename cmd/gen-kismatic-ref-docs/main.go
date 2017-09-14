package main

import (
	"flag"
	"fmt"
	"go/ast"
	godoc "go/doc"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strings"
)

var output = flag.String("o", "", "the output mode")

type doc struct {
	property     string
	propertyType string
	description  string
	defaultValue string
	options      []string
	required     bool
	deprecated   bool
}

func main() {
	flag.Parse()

	if len(flag.Args()) != 2 || *output == "" {
		fmt.Fprintf(os.Stderr, "usage: %s <path to go file> <type> -o <output type>\n", os.Args[0])
		os.Exit(1)
	}
	file := flag.Arg(0)
	typeName := flag.Arg(1)

	var r renderer
	switch *output {
	case "markdown":
		r = markdown{}
	case "markdown-table":
		r = markdownTable{}
	default:
		fmt.Fprintf(os.Stderr, "unknown output type: %s\n", *output)
		os.Exit(1)
	}

	fset := token.NewFileSet()
	m := make(map[string]*ast.File)

	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing file: %v\n", err)
		os.Exit(1)
	}

	m[file] = f
	apkg, _ := ast.NewPackage(fset, m, nil, nil) // error deliberately ignored
	pkgDoc := godoc.New(apkg, "", 0)

	for _, t := range pkgDoc.Types {
		if t.Name == typeName {
			docs := docForType(typeName, pkgDoc.Types, "")
			// Enforce non-empty documentation on all fields
			for _, d := range docs {
				if strings.TrimSpace(d.description) == "" {
					fmt.Fprintf(os.Stderr, "property %s does not have documentation\n", d.property)
					os.Exit(1)
				}
			}
			r.render(docs)
		}
	}

}

type renderer interface {
	render(docs []doc)
}

// performs a depth-first traversal of a type and returns a list of docs
func docForType(typeName string, allTypes []*godoc.Type, parentFieldName string) []doc {
	docs := []doc{}
	for _, t := range allTypes {
		if t.Name == typeName {
			typeSpec := t.Decl.Specs[0].(*ast.TypeSpec)
			switch tt := typeSpec.Type.(type) {
			default:
				panic(fmt.Sprintf("unhandled typespec type %s", typeSpec.Type))
			case *ast.Ident:
				// This case handles the type alias used for OptionalNodeGroup.
				// Recurse to get the docs for NodeGroup
				d := docForType(tt.Name, allTypes, parentFieldName)
				docs = append(docs, d...)
			case *ast.StructType:
				for _, f := range tt.Fields.List {
					fieldName := fieldName(parentFieldName, f)
					var typeName string

					// Figure out the type of the AST node
					switch x := f.Type.(type) {
					case *ast.StarExpr:
						typeName = x.X.(*ast.Ident).Name
						d, err := parseDoc(fieldName, typeName, f.Doc.Text())
						if err != nil {
							panic(err)
						}
						docs = append(docs, d)
						if isStruct(typeName) {
							docs = append(docs, docForType(typeName, allTypes, fieldName)...)
						}

					case *ast.Ident:
						typeName = x.Name
						d, err := parseDoc(fieldName, typeName, f.Doc.Text())
						if err != nil {
							panic(err)
						}
						docs = append(docs, d)
						if isStruct(typeName) {
							docs = append(docs, docForType(typeName, allTypes, fieldName)...)
						}

					// In the case of an array type, use []+typeName as the type
					// Recurse if it is an array of non-basic types.
					case *ast.ArrayType:
						typeName = x.Elt.(*ast.Ident).Name
						d, err := parseDoc(fieldName, "[]"+typeName, f.Doc.Text())
						if err != nil {
							panic(err)
						}
						docs = append(docs, d)
						if isStruct(typeName) {
							docs = append(docs, docForType(typeName, allTypes, fieldName)...)
						}
					default:
						panic(fmt.Sprintf("unhandled typespec type: %q", reflect.TypeOf(x).Name()))
					}
				}
			}
		}
	}
	return docs
}

func fieldName(parentFieldName string, field *ast.Field) string {
	var yamlTag string
	if field.Tag != nil {
		yamlTag = reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1]).Get("yaml") // Delete first and last quotation
	}
	// Get the field name from the yaml tag (e.g. allow_package_installation,omitempty)
	yamlTag = strings.Split(yamlTag, ",")[0]
	if yamlTag == "" {
		if field.Names != nil {
			yamlTag = strings.ToLower(field.Names[0].Name)
		} else {
			yamlTag = strings.ToLower(field.Type.(*ast.Ident).Name)
		}
	}
	if parentFieldName != "" {
		return parentFieldName + "." + yamlTag
	}
	return yamlTag
}

func parseDoc(propertyName string, propertyType string, typeDocs string) (doc, error) {
	d := doc{property: propertyName, propertyType: propertyType}
	lines := strings.Split(typeDocs, "\n")
	for _, l := range lines {
		if strings.Contains(l, "+required") {
			d.required = true
			continue
		}
		if strings.Contains(l, "+default") {
			d.defaultValue = strings.Split(l, "=")[1]
			continue
		}
		if strings.Contains(l, "+options") {
			optsString := strings.Split(l, "=")[1]
			d.options = strings.Split(optsString, ",")
			continue
		}
		if strings.Contains(l, "+deprecated") {
			d.deprecated = true
			continue
		}
		if strings.HasPrefix(l, "+") {
			return d, fmt.Errorf("unknown special marker found in documentation of %q. line was: %q", propertyName, l)
		}
		d.description = d.description + " " + l
	}
	return d, nil
}

func isStruct(s string) bool {
	return !basicTypes[s]
}

// not a comprehensive list, but works for now...
var basicTypes = map[string]bool{
	"bool":   true,
	"int":    true,
	"string": true,
}
