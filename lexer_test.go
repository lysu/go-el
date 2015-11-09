package el_test

import (
	"github.com/lysu/go-el"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLex(t *testing.T) {
	assert := assert.New(t)
	tok, err := el.Lex("abc.cc[1]")
	assert.Nil(err)
	assert.Len(tok, 6)
	assert.Equal("[", tok[3].Val)
	assert.Equal("]", tok[5].Val)
}
