package gen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"strings"
	"unicode"

	"github.com/iancoleman/strcase"
)

type GoWrapper struct {
	file *ast.File
	fset *token.FileSet

	structs map[string]*ast.StructType
	// this will hold a map of field identifiers to their json keys
	fieldNames map[string]string
}

func NewFromInput(input io.Reader) (*GoWrapper, error) {
	var generator GoWrapper

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "src", input, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("error generating ast file: %w", err)
	}

	generator.fset = fset
	generator.file = file
	generator.structs = make(map[string]*ast.StructType)
	generator.fieldNames = make(map[string]string)

	return &generator, nil
}

func (g *GoWrapper) extractStructs() {
	// loop through and fine type spec definitions
	for _, i := range g.file.Decls {
		d, ok := i.(*ast.GenDecl)
		if !ok || d.Tok != token.TYPE || len(d.Specs) != 1 {
			continue
		}

		// Now we are sure it a single type decleration
		spec, ok := d.Specs[0].(*ast.TypeSpec)
		if !ok {
			continue
		}
		t, ok := spec.Type.(*ast.StructType)
		if !ok {
			continue
		}
		g.structs[spec.Name.Name] = t
	}
}

func isCapitalized(name string) bool {
	runes := []rune(name)
	if len(runes) < 1 {
		return false
	}
	return unicode.IsUpper(runes[0])
}

func (g *GoWrapper) generateFieldNames() error {
	for name, s := range g.structs {
		for i, field := range s.Fields.List {
			if len(field.Names) != 1 {
				return fmt.Errorf("illegal field at %s", g.fset.Position(field.Pos()).String())
			}
			// If the field is not capitalized nilify(tm) it and skip
			if !isCapitalized(field.Names[0].Name) {
				s.Fields.List[i] = nil
				continue
			}

			fieldName := fmt.Sprintf("%s_%s", name, field.Names[0].Name)
			jsonName, err := g.getFieldJsonName(field)
			if err != nil {
				return err
			}

			g.fieldNames[fieldName] = jsonName
		}
	}
	return nil
}

func (g *GoWrapper) getFieldJsonName(field *ast.Field) (string, error) {
	if len(field.Names) != 1 {
		return "", fmt.Errorf("illegal field at %s", g.fset.Position(field.Pos()).String())
	}
	name := strcase.ToSnake(field.Names[0].Name)
	if field.Tag == nil {
		return name, nil
	}
	tags := strings.Split(field.Tag.Value, " ")
	if len(tags) < 1 {
		return name, nil
	}
	for _, tag := range tags {
		tag = strings.Trim(tag, "`")
		if !strings.HasPrefix(tag, "json:") {
			continue
		}
		tag = strings.TrimPrefix(tag, "json:")
		tag = strings.Trim(tag, "\"")
		tagValues := strings.Split(tag, ",")
		if len(tagValues) < 1 {
			return "", fmt.Errorf("invalid  json tag values at %s", g.fset.Position(field.Pos()).String())
		}
		if tagValues[0] != "" {
			name = tagValues[0]
		}
	}
	return name, nil
}

func (g *GoWrapper) GenerateJsonAst() (*JSONWrapper, error) {
	//	var rootNodes []*jsonast.RootNode
	//	for name, s := range g.structs {
	//		root := jsonast.RootNode{
	//			Type: jsonast.ObjectRoot,
	//		}
	//	}

	return nil, nil
}
