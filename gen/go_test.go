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

	fmt.Println(obj)
}
