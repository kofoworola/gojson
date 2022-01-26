package gen

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractStructs(t *testing.T) {
	file, err := os.Open("simple_test.txt")
	assert.Nil(t, err, "error opening file")
	defer file.Close()

	wrapper, err := NewFromInput(file)
	assert.Nil(t, err, "error creating new wrapper")
	wrapper.extractStructs()

	assert.Len(t, wrapper.structs, 1)

	if err := wrapper.generateFieldNames(); err != nil {
		t.Fatal(err)
	}

	assert.Len(t, wrapper.fieldNames, 2)
	assert.Equal(t, "test", wrapper.fieldNames["test_Name"])
	assert.Equal(t, "pass_is", wrapper.fieldNames["test_PassIs"])
}
