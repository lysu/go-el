package patcher_test

import (
	"github.com/lysu/go-struct-patcher"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLex(t *testing.T) {
	assert := assert.New(t)
	tok, err := patcher.Lex("abc.cc[1]")
	assert.Nil(err)
	assert.Len(tok, 6)
	assert.Equal("[", tok[3].Val)
	assert.Equal("]", tok[5].Val)
}
