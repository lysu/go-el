package patcher

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

	value, err := exp.Evaluate(target)

	if err != nil {
		return nil, err
	}

	return value, nil

}
