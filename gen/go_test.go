package gen

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractStructs(t *testing.T) {
	file, err := os.Open("simple_test.txt")
	defer file.Close()
	assert.Nil(t, err, "error opening file")
	defer file.Close()

	dat, err := io.ReadAll(file)
	assert.Nil(t, err)

	wrapper, err := NewFromString(string(dat))
	obj, err := wrapper.GenerateJSONAst()
	assert.Nil(t, err, "error creating new wrapper")

	for _, o := range obj {
		fmt.Printf("value is:\n%s\n\n", o.String(0))
	}
}
