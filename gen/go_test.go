package gen

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractStructs(t *testing.T) {
	file, err := os.Open("simple_test.txt")
	assert.Nil(t, err, "error opening file")
	defer file.Close()

	wrapper, err := NewFromInput(file)
	obj, err := wrapper.GenerateJSONAst()
	assert.Nil(t, err, "error creating new wrapper")

	for _, o := range obj {
		fmt.Printf("value is:\n%s\n\n", o.String(0))
	}
}
