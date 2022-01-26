package gen

import (
	"fmt"
	"io"

	"github.com/bradford-hamilton/dora/pkg/ast"
	"github.com/bradford-hamilton/dora/pkg/lexer"
	"github.com/bradford-hamilton/dora/pkg/parser"
)

type JSONWrapper struct {
	ast *ast.RootNode
}

func NewJsonFromInput(i io.Reader) (*JSONWrapper, error) {
	b, err := io.ReadAll(i)
	if err != nil {
		return nil, err
	}

	input := string(b)
	tree, err := parser.New(lexer.New(input)).ParseJSON()
	if err != nil {
		return nil, fmt.Errorf("error parsing json: %w", err)
	}

	return &JSONWrapper{
		ast: &tree,
	}, nil
}
