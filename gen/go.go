package gen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math/rand"
	"strings"
	"time"
	"unicode"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/iancoleman/strcase"
	"github.com/kofoworola/gojson/jsonast"
)

type GoWrapper struct {
	file *ast.File
	fset *token.FileSet

	structs            map[string]*ast.StructType
	orderedStructNames []string
	// this will hold a map of field identifiers to their json keys
	fieldNames       map[string]string
	generatedStructs map[string]*jsonast.Object
}

func NewFromString(input string) (*GoWrapper, error) {
	var generator GoWrapper

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "src", input, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("error generating ast file: %w", err)
	}

	generator.fset = fset
	generator.file = file
	generator.orderedStructNames = make([]string, 0)
	generator.structs = make(map[string]*ast.StructType)
	generator.fieldNames = make(map[string]string)
	generator.generatedStructs = make(map[string]*jsonast.Object)

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
		g.orderedStructNames = append(g.orderedStructNames, spec.Name.Name)
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

func (g *GoWrapper) GenerateJSONAst() ([]*jsonast.Object, error) {
	g.extractStructs()
	if err := g.generateFieldNames(); err != nil {
		return nil, err
	}
	var objects []*jsonast.Object
	for _, name := range g.orderedStructNames {
		s := g.structs[name]
		objects = append(objects, g.structToJsonObject(name, s))
	}
	return objects, nil
}

func (g *GoWrapper) structToJsonObject(name string, str *ast.StructType) *jsonast.Object {
	if res, ok := g.generatedStructs[name]; ok {
		return res
	}

	properties := make(map[string]jsonast.Value)
	for _, field := range str.Fields.List {
		if field == nil {
			continue
		}
		fieldName := fmt.Sprintf("%s_%s", name, field.Names[0].Name)
		jsonName := g.fieldNames[fieldName]
		if jsonName == "" {
			name, err := g.getFieldJsonName(field)
			if err != nil {
				panic(err)
			}

			jsonName = name
		}

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
					Type:  jsonast.StringType,
					Value: randStringGenerator(jsonName),
				}
			case "bool":
				properties[jsonName] = &jsonast.Literal{
					Type:  jsonast.BoolType,
					Value: true,
				}
			}
		}
		if arr, ok := field.Type.(*ast.ArrayType); ok {
			properties[jsonName] = g.arrayToJsonObject(arr)
		}
		if str, ok := field.Type.(*ast.StructType); ok {
			properties[jsonName] = g.structToJsonObject(field.Names[0].Name, str)
		}
	}
	obj := &jsonast.Object{
		Key:      name,
		Children: properties,
	}
	g.generatedStructs[name] = obj
	return obj
}

func (g *GoWrapper) arrayToJsonObject(arr *ast.ArrayType) *jsonast.Array {
	children := make([]jsonast.Value, 0)
	if ident, ok := arr.Elt.(*ast.Ident); ok {
		t := ident.Name
		var funcGen func() interface{}
		var jsonType jsonast.LiteralType
		switch t {
		case "int":
			funcGen = func() interface{} { return randNumber() }
			jsonType = jsonast.IntegerType
		case "string":
			funcGen = func() interface{} { return randStringGenerator("") }
			jsonType = jsonast.StringType
		case "bool":
			funcGen = func() interface{} { return true }
			jsonType = jsonast.BoolType
		}
		if t == "int" || t == "string" || t == "bool" {
			for i := 0; i < 5; i++ {
				children = append(
					children,
					&jsonast.Literal{
						Type:  jsonType,
						Value: funcGen(),
					},
				)
			}
		}
		if str, ok := g.structs[t]; ok {
			jsonObj := g.structToJsonObject(t, str)
			children = append(children, jsonObj)
		}
	}
	if str, ok := arr.Elt.(*ast.StructType); ok {
		jsonObj := g.structToJsonObject("", str)
		for i := 0; i < 2; i++ {
			children = append(children, jsonObj)
		}

	}

	return &jsonast.Array{
		Children: children,
	}

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
	case "country":
		val = gofakeit.Country()
	default:
		val = gofakeit.SentenceSimple()
	}

	if strings.HasSuffix(fieldName, "address") {
		val = gofakeit.Address().Address
	}
	if strings.HasSuffix(fieldName, "phone") {
		val = gofakeit.Phone()
	}

	return val
}
