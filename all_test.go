package patcher_test

import (
	"testing"

	"github.com/lysu/go-struct-patcher"
	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {

	assert := assert.New(t)

	testPath := "abc.r1.d"

	tok, err := patcher.Lex(testPath)
	assert.Nil(err)

	assert.Len(tok, 5)

	patcher := patcher.NewParser(tok)
	assert.NotNil(patcher)

}
