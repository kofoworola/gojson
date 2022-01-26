package jsonast

import (
	"fmt"
	"strings"
)

type LiteralType int

const (
	StringType LiteralType = iota
	IntegerType
	BoolType
)

const INDENTSTRING = "\t"

type Value interface {
	isValue()
	String(indentCount int) string
}

type Object struct {
	Key      string
	Children map[string]Value
}

func (o *Object) isValue() {}

func (o *Object) String(indentCount int) string {
	builder := strings.Builder{}
	builder.WriteString("{")
	for name, c := range o.Children {
		builder.WriteString(fmt.Sprintf(
			"\n%s\"%s\": %s,",
			strings.Repeat(INDENTSTRING, indentCount+1),
			name,
			c.String(indentCount+1),
		))
	}
	gen := strings.TrimSuffix(builder.String(), ",")
	return fmt.Sprintf("%s\n%s}", gen, strings.Repeat(INDENTSTRING, indentCount))
}

type Array struct {
	Children []Value
}

func (a *Array) isValue() {}

func (a *Array) String(indentCount int) string {
	builder := strings.Builder{}
	builder.WriteString("[")
	for _, c := range a.Children {
		builder.WriteString(fmt.Sprintf(
			"\n%s%s,",
			strings.Repeat(INDENTSTRING, indentCount+1),
			c.String(indentCount+1),
		))
	}
	gen := strings.TrimSuffix(builder.String(), ",")
	return fmt.Sprintf("%s\n%s]", gen, strings.Repeat(INDENTSTRING, indentCount))

}

type Literal struct {
	Type  LiteralType
	Value interface{}
}

func (l *Literal) isValue() {}

func (l *Literal) String(indentCount int) string {
	switch l.Type {
	case StringType:
		return fmt.Sprintf(`"%s"`, l.Value.(string))
	case IntegerType:
		return fmt.Sprintf(`%d`, l.Value.(int64))
	case BoolType:
		v := l.Value.(bool)
		if v {
			return "true"
		} else {
			return "false"
		}
	default:
		return ""
	}
}
