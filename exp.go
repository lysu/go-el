package patcher

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	varTypeInt = iota
	varTypeIdent
)

type IEvaluator interface {
	GetPositionToken() *Token
	Evaluate(target interface{}) (*Value, *Error)
}

type intResolver struct {
	locationToken *Token
	val           int
}

func (i *intResolver) Evaluate(target interface{}) (*Value, *Error) {
	return AsValue(i.val), nil
}

func (i *intResolver) GetPositionToken() *Token {
	return i.locationToken
}

type stringResolver struct {
	locationToken *Token
	val           string
}

func (s *stringResolver) Evaluate(target interface{}) (*Value, *Error) {
	return AsValue(s.val), nil
}

func (s *stringResolver) GetPositionToken() *Token {
	return s.locationToken
}

type boolResolver struct {
	locationToken *Token
	val           bool
}

func (b *boolResolver) Evaluate(target interface{}) (*Value, *Error) {
	return AsValue(b.val), nil
}

func (b *boolResolver) GetPositionToken() *Token {
	return b.locationToken
}

type variableResolver struct {
	locationToken *Token

	parts []*variablePart
}

type functionCallArgument interface {
	Evaluate(target interface{}) (*Value, *Error)
}

func (vr *variableResolver) Evaluate(target interface{}) (*Value, *Error) {
	value, err := vr.resolve(target)
	if err != nil {
		return AsValue(nil), NewError(err.Error(), vr.locationToken)
	}
	return value, nil
}

func (vr *variableResolver) String() string {
	parts := make([]string, 0, len(vr.parts))
	for _, p := range vr.parts {
		switch p.typ {
		case varTypeInt:
			parts = append(parts, strconv.Itoa(p.i))
		case varTypeIdent:
			parts = append(parts, p.s)
		default:
			panic("unimplemented")
		}
	}
	return strings.Join(parts, ".")
}

func (vr *variableResolver) resolve(target interface{}) (*Value, error) {
	current := reflect.ValueOf(target)

	for _, part := range vr.parts {
		// Before resolving the pointer, let's see if we have a method to call
		// Problem with resolving the pointer is we're changing the receiver
		isFunc := false
		if part.typ == varTypeIdent {
			funcValue := current.MethodByName(part.s)
			if funcValue.IsValid() {
				current = funcValue
				isFunc = true
			}
		}

		if !isFunc {
			// If current a pointer, resolve it
			if current.Kind() == reflect.Ptr {
				current = current.Elem()
				if !current.IsValid() {
					// Value is not valid (anymore)
					return AsValue(nil), nil
				}
			}

			// Look up which part must be called now
			switch part.typ {
			case varTypeInt:
				// Calling an index is only possible for:
				// * slices/arrays/strings
				switch current.Kind() {
				case reflect.String, reflect.Array, reflect.Slice:
					if current.Len() > part.i {
						current = current.Index(part.i)
					} else {
						return nil, fmt.Errorf("Index out of range: %d (variable %s)", part.i, vr.String())
					}
				default:
					return nil, fmt.Errorf("Can't access an index on type %s (variable %s)",
						current.Kind().String(), vr.String())
				}
			case varTypeIdent:
				// debugging:
				// fmt.Printf("now = %s (kind: %s)\n", part.s, current.Kind().String())

				// Calling a field or key
				switch current.Kind() {
				case reflect.Struct:
					current = current.FieldByName(part.s)
				case reflect.Map:
					current = current.MapIndex(reflect.ValueOf(part.s))
				default:
					return nil, fmt.Errorf("Can't access a field by name on type %s (variable %s)",
						current.Kind().String(), vr.String())
				}
			default:
				panic("unimplemented")
			}
		}

		if !current.IsValid() {
			// Value is not valid (anymore)
			return AsValue(nil), nil
		}

		// If current is a reflect.ValueOf(pongo2.Value), then unpack it
		// Happens in function calls (as a return value) or by injecting
		// into the execution context (e.g. in a for-loop)
		if current.Type() == reflect.TypeOf(&Value{}) {
			tmpValue := current.Interface().(*Value)
			current = tmpValue.val
		}

		// Check whether this is an interface and resolve it where required
		if current.Kind() == reflect.Interface {
			current = reflect.ValueOf(current.Interface())
		}

		// Check if the part is a function call
		if part.isFunctionCall || current.Kind() == reflect.Func {
			// Check for callable
			if current.Kind() != reflect.Func {
				return nil, fmt.Errorf("'%s' is not a function (it is %s)", vr.String(), current.Kind().String())
			}

			// Check for correct function syntax and types
			// func(*Value, ...) *Value
			t := current.Type()

			// Input arguments
			if len(part.callingArgs) != t.NumIn() && !(len(part.callingArgs) >= t.NumIn()-1 && t.IsVariadic()) {
				return nil,
					fmt.Errorf("Function input argument count (%d) of '%s' must be equal to the calling argument count (%d).",
						t.NumIn(), vr.String(), len(part.callingArgs))
			}

			// Output arguments
			if t.NumOut() != 1 {
				return nil, fmt.Errorf("'%s' must have exactly 1 output argument", vr.String())
			}

			// Evaluate all parameters
			var parameters []reflect.Value

			numArgs := t.NumIn()
			isVariadic := t.IsVariadic()
			var fnArg reflect.Type

			for idx, arg := range part.callingArgs {
				pv, err := arg.Evaluate(target)
				if err != nil {
					return nil, err
				}

				if isVariadic {
					if idx >= t.NumIn()-1 {
						fnArg = t.In(numArgs - 1).Elem()
					} else {
						fnArg = t.In(idx)
					}
				} else {
					fnArg = t.In(idx)
				}

				if fnArg != reflect.TypeOf(new(Value)) {
					// Function's argument is not a *pongo2.Value, then we have to check whether input argument is of the same type as the function's argument
					if !isVariadic {
						if fnArg != reflect.TypeOf(pv.Interface()) && fnArg.Kind() != reflect.Interface {
							return nil, fmt.Errorf("Function input argument %d of '%s' must be of type %s or *pongo2.Value (not %T).",
								idx, vr.String(), fnArg.String(), pv.Interface())
						}
						// Function's argument has another type, using the interface-value
						parameters = append(parameters, reflect.ValueOf(pv.Interface()))
					} else {
						if fnArg != reflect.TypeOf(pv.Interface()) && fnArg.Kind() != reflect.Interface {
							return nil, fmt.Errorf("Function variadic input argument of '%s' must be of type %s or *pongo2.Value (not %T).",
								vr.String(), fnArg.String(), pv.Interface())
						}
						// Function's argument has another type, using the interface-value
						parameters = append(parameters, reflect.ValueOf(pv.Interface()))
					}
				} else {
					// Function's argument is a *pongo2.Value
					parameters = append(parameters, reflect.ValueOf(pv))
				}
			}

			// Check if any of the values are invalid
			for _, p := range parameters {
				if p.Kind() == reflect.Invalid {
					return nil, fmt.Errorf("Calling a function using an invalid parameter")
				}
			}

			// Call it and get first return parameter back
			rv := current.Call(parameters)[0]

			if rv.Type() != reflect.TypeOf(new(Value)) {
				current = reflect.ValueOf(rv.Interface())
			} else {
				// Return the function call value
				current = rv.Interface().(*Value).val
			}
		}
	}

	if !current.IsValid() {
		// Value is not valid (e. g. NIL value)
		return AsValue(nil), nil
	}

	return &Value{val: current}, nil
}

func (vr *variableResolver) GetPositionToken() *Token {
	return vr.locationToken
}

type variablePart struct {
	typ int
	s   string
	i   int

	isFunctionCall bool
	callingArgs    []functionCallArgument // needed for a function call, represents all argument nodes (INode supports nested function calls)
}

func (p *Parser) ParseExp() (IEvaluator, *Error) {

	if p.Match(TokenSymbol, "(") != nil {
		expr, err := p.ParseExp()
		if err != nil {
			return nil, err
		}
		if p.Match(TokenSymbol, ")") == nil {
			return nil, p.Error("Closing bracket expected after expression", nil)
		}
		return expr, nil
	}

	t := p.Current()

	if t == nil {
		return nil, p.Error("Unexpect EOF, expected an identifier", p.lastToken)
	}

	switch t.Typ {
	case TokenNumber:
		p.Consume()
		i, err := strconv.Atoi(t.Val)
		if err != nil {
			return nil, p.Error(err.Error(), t)
		}
		nr := &intResolver{
			locationToken: t,
			val:           i,
		}
		return nr, nil
	case TokenString:
		p.Consume()
		sr := &stringResolver{
			locationToken: t,
			val:           t.Val,
		}
		return sr, nil
	case TokenKeyword:
		p.Consume()
		switch t.Val {
		case "true":
			br := &boolResolver{
				locationToken: t,
				val:           true,
			}
			return br, nil
		case "false":
			br := &boolResolver{
				locationToken: t,
				val:           false,
			}
			return br, nil
		default:
			return nil, p.Error("This keyword is not allowed here.", nil)
		}
	}

	if t.Typ != TokenIdentifier {
		return nil, p.Error("Expected either a number, string, keyword or identifier.", t)
	}

	resolver := &variableResolver{
		locationToken: t,
	}

	resolver.parts = append(resolver.parts, &variablePart{
		typ: varTypeIdent,
		s:   t.Val,
	})

	p.Consume()

variableLoop:
	for p.Remaining() > 0 {
		t = p.Current()

		if p.Match(TokenSymbol, ".") != nil {
			t2 := p.Current()
			if t2 != nil {
				switch t2.Typ {
				case TokenIdentifier:
					resolver.parts = append(resolver.parts, &variablePart{
						typ: varTypeIdent,
						s:   t2.Val,
					})
					p.Consume()
					continue variableLoop
				case TokenNumber:
					i, err := strconv.Atoi(t2.Val)
					if err != nil {
						return nil, p.Error(err.Error(), t2)
					}
					resolver.parts = append(resolver.parts, &variablePart{
						typ: varTypeInt,
						i:   i,
					})
					p.Consume()
					continue variableLoop
				default:
					return nil, p.Error("This token is not allowed within a variable name", t2)
				}
			} else {
				return nil, p.Error("Unexpected EOF", p.lastToken)
			}
		} else if p.Match(TokenSymbol, "(") != nil {
			// Function call
			// FunctionName '(' Comma-separated list of expressions ')'
			part := resolver.parts[len(resolver.parts)-1]
			part.isFunctionCall = true
		argumentLoop:
			for {
				if p.Remaining() == 0 {
					return nil, p.Error("Unexpected EOF, expected function call argument list.", p.lastToken)
				}

				if p.Peek(TokenSymbol, ")") == nil {
					// No closing bracket, so we're parsing an expression
					exprArg, err := p.ParseExp()
					if err != nil {
						return nil, err
					}
					part.callingArgs = append(part.callingArgs, exprArg)

					if p.Match(TokenSymbol, ")") != nil {
						// If there's a closing bracket after an expression, we will stop parsing the arguments
						break argumentLoop
					} else {
						// If there's NO closing bracket, there MUST be an comma
						if p.Match(TokenSymbol, ",") == nil {
							return nil, p.Error("Missing comma or closing bracket after argument.", nil)
						}
					}
				} else {
					// We got a closing bracket, so stop parsing arguments
					p.Consume()
					break argumentLoop
				}

			}
			// We're done parsing the function call, next variable part
			continue variableLoop
		}

		break

	}
	return resolver, nil
}
