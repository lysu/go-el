package el_test

import (
	"testing"

	"encoding/json"
	"github.com/lysu/go-el"
	"github.com/stretchr/testify/assert"
	"strconv"
)

type User struct {
	Name      string
	BizState  map[string]int
	ImgIDList []int
	Images    []*Image
	ImgIdx    map[string]*Image
}

type Image struct {
	Content string
}

func (t User) FindImage(i int) *Image {
	return t.Images[i]
}

func (t User) LocateImage(i int) *Image {
	return t.ImgIdx[strconv.Itoa(i)]
}

func TestLocate(t *testing.T) {

	data := User{
		Name:      "ほん",
		ImgIDList: []int{0, 1, 2},
		Images:    []*Image{{"1.jpg"}, {"2.jpg"}, {"3.jpg"}},
		ImgIdx: map[string]*Image{
			"0": {"しゃしん１.jpg"},
			"1": {"写真2.jpg"},
			"2": {"写真3.jpg"},
		},
	}

	exp := el.Expression("Name")
	v, err := exp.Execute(&data)
	assert.NoError(t, err)
	err = v.SetValue("zzzz")
	assert.NoError(t, err)
	assert.Equal(t, "zzzz", data.Name)

	exp = el.Expression("ImgIDList.0")
	v, err = exp.Execute(&data)
	assert.NoError(t, err)
	err = v.SetValue(json.Number("9"))
	assert.NoError(t, err)
	assert.Equal(t, 9, data.ImgIDList[0])

	exp = el.Expression("ImgIDList[0]")
	v, err = exp.Execute(&data)
	assert.NoError(t, err)
	assert.Equal(t, 9, data.ImgIDList[0])

	exp = el.Expression("Images[ImgIDList[2]].Content")
	v, err = exp.Execute(&data)
	assert.NoError(t, err)
	err = v.SetValue("しゃ")
	assert.NoError(t, err)
	assert.Equal(t, "しゃ", v.String())

}

func TestFunctionCall(t *testing.T) {

	data := User{
		Name:      "ほん",
		ImgIDList: []int{0, 1, 2},
		Images:    []*Image{{"1.jpg"}, {"2.jpg"}, {"3.jpg"}},
		ImgIdx: map[string]*Image{
			"0": {"しゃしん１.jpg"},
			"1": {"写真2.jpg"},
			"2": {"写真3.jpg"},
		},
	}

	exp := el.Expression("FindImage(ImgIDList.1).Content")
	v, err := exp.Execute(&data)
	assert.NoError(t, err)
	err = v.SetValue("なに")
	assert.NoError(t, err)
	assert.Equal(t, "なに", data.Images[1].Content)

	exp = el.Expression("LocateImage(ImgIDList.2).Content")
	v, err = exp.Execute(&data)
	assert.NoError(t, err)
	err = v.SetValue("なん")
	assert.NoError(t, err)
	assert.Equal(t, "なん", data.ImgIdx["2"].Content)

}

func TestIndexAccess(t *testing.T) {

	user := User{
		Name:      "ほん",
		ImgIDList: []int{0, 1, 2},
		Images:    []*Image{{"1.jpg"}, {"2.jpg"}, {"3.jpg"}},
		ImgIdx: map[string]*Image{
			"0": {"しゃしん１.jpg"},
			"1": {"しゃしん2.jpg"},
			"2": {"しゃしん3.jpg"},
		},
	}

	exp := el.Expression("ImgIDList[2]")
	v, err := exp.Execute(&user)
	assert.NoError(t, err)
	assert.Equal(t, 2, v.Integer())
	err = v.SetValue(7)
	assert.NoError(t, err)
	assert.Equal(t, 7, user.ImgIDList[2])

	exp = el.Expression("ImgIdx[2].Content")
	v, err = exp.Execute(&user)
	assert.NoError(t, err)
	assert.Equal(t, "しゃしん3.jpg", v.String())
	err = v.SetValue("しゃしん4.jpg")
	assert.NoError(t, err)
	assert.Equal(t, "しゃしん4.jpg", user.ImgIdx["2"].Content)

	exp = el.Expression("ImgIdx[ImgIDList[0]].Content")
	v, err = exp.Execute(&user)
	assert.NoError(t, err)
	assert.Equal(t, "しゃしん１.jpg", v.String())
	err = v.SetValue("しゃしん233.jpg")
	assert.NoError(t, err)
	assert.Equal(t, "しゃしん233.jpg", user.ImgIdx["0"].Content)

}

func TestIndexSet(t *testing.T) {
	user := User{
		Name:      "ほん",
		ImgIDList: []int{0, 1, 2},
		Images:    []*Image{{"1.jpg"}, {"2.jpg"}, {"3.jpg"}},
		ImgIdx: map[string]*Image{
			"0": {"しゃしん１.jpg"},
			"1": {"しゃしん2.jpg"},
			"2": {"しゃしん3.jpg"},
		},
		BizState: map[string]int{},
	}
	exp := el.Expression("BizState[3]")
	v, err := exp.Execute(&user)
	assert.NoError(t, err)
	assert.True(t, v.IsNil())
	err = v.SetValue(3)
	assert.NoError(t, err)
	assert.Equal(t, 3, user.BizState["3"])

	exp = el.Expression("ImgIDList[99]")
	v, err = exp.Execute(&user)
	assert.Nil(t, err)
	v.SetValue(99)
	assert.Equal(t, 99, user.ImgIDList[99])
}
