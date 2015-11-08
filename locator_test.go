package patcher_test

import (
	"fmt"
	"testing"

	"github.com/lysu/go-struct-patcher"
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

func TestLocate(t *testing.T) {

	p := patcher.Path("abc.Index(abc.Idx.2).Content")

	data := T{
		Idx: []int{1, 2, 3},
		B:   "Yue",
	}

	v, err := patcher.Locate(data, p)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s-------\n", v.String())

}
