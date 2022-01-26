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

type Value interface {
	isValue()
	String() string
}

type Object struct {
	Key      string
	Children map[string]Value
}

func (o *Object) isValue() {}

func (o *Object) String() string {
	builder := strings.Builder{}
	builder.WriteRune('{')
	for name, c := range o.Children {
		builder.WriteString(fmt.Sprintf(`"%s":%s,`, name, c.String()))
	}
	gen := strings.TrimSuffix(builder.String(), ",")
	return gen + "}"
}

type Array struct {
	Children []Value
}

func (a *Array) isValue() {}

func (a *Array) String() string {
	builder := strings.Builder{}
	builder.WriteRune('[')
	for i, c := range a.Children {
		builder.WriteString(c.String())
		if i != len(a.Children)-1 {
			builder.WriteRune(',')
		}
	}
	builder.WriteRune(']')

	return builder.String()

}

type Literal struct {
	Type  LiteralType
	Value interface{}
}

func (l *Literal) isValue() {}

func (l *Literal) String() string {
	switch l.Type {
	case StringType:
		return fmt.Sprintf(`"%s"`, l.Value.(string))
	case IntegerType:
		return fmt.Sprintf(`"%d"`, l.Value.(int64))
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
