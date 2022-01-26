package jsonast

type LiteralType int

const (
	StringType LiteralType = iota
	IntegerType
	BoolType
)

type Value interface {
	isValue()
}

type Object struct {
	Key      string
	Children map[string]Value
}

func (o *Object) isValue() {}

type Array struct {
	Children []Value
}

func (a *Array) isValue() {}

type Literal struct {
	Type  LiteralType
	Value interface{}
}

func (l *Literal) isValue() {}
