package patcher

const ContextKey = "t"
const DotSymbol = "."

func Locate(target interface{}, path Path) (*Value, error) {

	toks, err := Lex(string(path))
	if err != nil {
		return nil, err
	}

	parser := NewParser(toks)

	exp, err := parser.ParseExp()
	if err != nil {
		return nil, err
	}

	err = exp.Execute(target)

	if err != nil {
		return nil, err
	}

	return nil, nil

}
