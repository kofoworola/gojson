package gen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"math/rand"
	"strings"
	"time"
	"unicode"

	"github.com/brianvoe/gofakeit"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/iancoleman/strcase"
	"github.com/kofoworola/gojson/jsonast"
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

			// quickly check if the field is an embedded struct and add to list
			// or if the field is an array of embedded struct
			if strType, ok := field.Type.(*ast.StructType); ok {
				g.structs[field.Names[0].Name] = strType
			}

			if arrType, ok := field.Type.(*ast.ArrayType); ok {
				if strType, ok := arrType.Elt.(*ast.StructType); ok {
					g.structs[field.Names[0].Name] = strType
				}
			}

			// use the format structname_fieldname as the field name key
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

//func (g *GoWrapper) GenerateJSONAst() (*jsonast.Object, error) {
//	var objects []*jsonast.Object
//	for name, s := range g.structs {
//		properties := make(map[string]jsonast.Value)
//	}
//	return nil, nil
//}

func (g *GoWrapper) structToJsonObject(name string, str *ast.StructType) *jsonast.Object {
	properties := make(map[string]jsonast.Value)
	for _, field := range str.Fields.List {
		fieldName := fmt.Sprintf("%s_%s", name, field.Names[0].Name)
		jsonName := g.fieldNames[fieldName]
		if ident, ok := field.Type.(*ast.Ident); ok {
			// check if it is a string, int, bool, custom struct identifier
			if val, ok := g.structs[ident.Name]; ok {
				properties[jsonName] = g.structToJsonObject(ident.Name, val)
				continue
			}
			switch ident.Name {
			case "int":
				properties[jsonName] = &jsonast.Literal{
					Type:  jsonast.IntegerType,
					Value: randNumber(),
				}
			case "string":
				properties[jsonName] = &jsonast.Literal{
					Type:  jsonast.IntegerType,
					Value: "",
				}
			}
		}
	}
	return nil
}

func randNumber() int64 {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	return r.Int63()
}

func randStringGenerator(fieldName string) string {
	var val string
	switch fieldName {
	case "email":
		val = gofakeit.Email()
	case "name", "first_name", "last_name", "middle_name":
		val = gofakeit.Name()
	case "company":
		val = gofakeit.Company()
	default:
		val = gofakeit.Name()
	}

	if strings.HasSuffix(fieldName, "address") {
		val = gofakeit.Address().Address()
	}
	if strings.HasSuffix(fieldName, "phone") {
		val = gofakeit.Phone()
	}

	return val
}
