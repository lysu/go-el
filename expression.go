package el

import "strings"

// Expression to Patch
type Expression string

func (path *Expression) Execute(target interface{}) (*Value, error) {

	toks, err := Lex(string(*path))
	if err != nil {
		return nil, err
	}

	parser := NewParser(toks)

	exp, err := parser.ParseExp()
	if err != nil {
		return nil, err
	}

	value, err := exp.Evaluate(target)

	if err != nil {
		return nil, err
	}

	return value, nil

}

func (p Expression) FirstPart() string {
	idx := strings.Index(string(p), ".")
	if idx == -1 {
		return ""
	}
	return upperFirst(string(p)[:idx])
}
