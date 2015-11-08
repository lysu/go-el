package patcher_test

import (
	"testing"

	"encoding/json"
	"github.com/lysu/go-struct-patcher"
	"github.com/stretchr/testify/assert"
)

type User struct {
	Name      string
	ImgIDList []int
	Images    []*Image
}

type Image struct {
	Content string
}

func (t User) FindImage(i int) *Image {
	return t.Images[i]
}

func TestLocate(t *testing.T) {

	data := User{
		Name:      "ほん",
		ImgIDList: []int{0, 1, 2},
		Images:    []*Image{{"1.jpg"}, {"2.jpg"}, {"3.jpg"}},
	}

	v, err := patcher.Locate(&data, patcher.Path("Name"))
	assert.NoError(t, err)

	err = v.SetValue("zzzz")
	assert.NoError(t, err)
	assert.Equal(t, "zzzz", data.Name)

	v, err = patcher.Locate(&data, patcher.Path("ImgIDList.0"))
	assert.NoError(t, err)
	err = v.SetValue(json.Number("9"))
	assert.NoError(t, err)

	assert.Equal(t, 9, data.ImgIDList[0])

	v, err = patcher.Locate(&data, patcher.Path("FindImage(ImgIDList.1).Content"))
	assert.NoError(t, err)

	err = v.SetValue("なに")
	assert.NoError(t, err)
	assert.Equal(t, "なに", data.Images[1].Content)

}
