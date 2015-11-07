package patcher_test

import (
	"fmt"
	"testing"

	"github.com/lysu/go-struct-patcher"
	"github.com/stretchr/testify/assert"
)

type T struct {
	Idx []int
	B   string
}

type S struct {
	Content string
}

func (t T) Index(i int) S {
	return S{
		Content: fmt.Sprintf("abc---%d", i),
	}
}

func TestAll(t *testing.T) {

	assert := assert.New(t)

	testPath := "abc.Index(abc.Idx.2).Content"

	tok, err := patcher.Lex(testPath)
	assert.Nil(err)

	assert.NotEmpty(tok)

	parser := patcher.NewParser(tok)
	assert.NotNil(parser)

	v, err := parser.ParseVar()
	assert.Nil(err)
	assert.NotNil(v)

	data := T{
		Idx: []int{1, 2, 3},
		B:   "Yue",
	}

	v.Execute(patcher.NewExecutionContext(patcher.Context{
		"abc": data,
	}))

}
