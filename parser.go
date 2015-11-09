package el

type Parser struct {
	idx       int
	tokens    []*Token
	lastToken *Token
}

func NewParser(tokens []*Token) *Parser {
	p := &Parser{tokens: tokens}
	if len(tokens) > 0 {
		p.lastToken = tokens[len(tokens)-1]
	}
	return p
}

func (p *Parser) Consume() {
	p.ConsumeN(1)
}

func (p *Parser) ConsumeN(count int) {
	p.idx += count
}

func (p *Parser) Current() *Token {
	return p.Get(p.idx)
}

func (p *Parser) MatchType(typ TokenType) *Token {
	if t := p.PeekType(typ); t != nil {
		p.Consume()
		return t
	}
	return nil
}

func (p *Parser) Match(typ TokenType, val string) *Token {
	if t := p.Peek(typ, val); t != nil {
		p.Consume()
		return t
	}
	return nil
}

func (p *Parser) MatchOne(typ TokenType, vals ...string) *Token {
	for _, val := range vals {
		if t := p.Peek(typ, val); t != nil {
			p.Consume()
			return t
		}
	}
	return nil
}

func (p *Parser) PeekType(typ TokenType) *Token {
	return p.PeekTypeN(0, typ)
}

func (p *Parser) Peek(typ TokenType, val string) *Token {
	return p.PeekN(0, typ, val)
}

func (p *Parser) PeekOne(typ TokenType, vals ...string) *Token {
	for _, v := range vals {
		t := p.PeekN(0, typ, v)
		if t != nil {
			return t
		}
	}
	return nil
}

func (p *Parser) PeekN(shift int, typ TokenType, val string) *Token {
	t := p.Get(p.idx + shift)
	if t != nil {
		if t.Typ == typ && t.Val == val {
			return t
		}
	}
	return nil
}

func (p *Parser) PeekTypeN(shift int, typ TokenType) *Token {
	t := p.Get(p.idx + shift)
	if t != nil {
		if t.Typ == typ {
			return t
		}
	}
	return nil
}

func (p *Parser) Remaining() int {
	return len(p.tokens) - p.idx
}

func (p *Parser) Count() int {
	return len(p.tokens)
}

func (p *Parser) Get(i int) *Token {
	if i < len(p.tokens) {
		return p.tokens[i]
	}
	return nil
}

func (p *Parser) GetR(shift int) *Token {
	i := p.idx + shift
	return p.Get(i)
}

func (p *Parser) Error(msg string, token *Token) *Error {
	if token == nil {
		// Set current token
		token = p.Current()
		if token == nil {
			// Set to last token
			if len(p.tokens) > 0 {
				token = p.tokens[len(p.tokens)-1]
			}
		}
	}
	var line, col int
	if token != nil {
		line = token.Line
		col = token.Col
	}
	return &Error{
		Line:     line,
		Column:   col,
		Token:    token,
		ErrorMsg: msg,
	}
}
