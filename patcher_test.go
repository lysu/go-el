package patcher_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/lysu/go-struct-patcher"
	"github.com/stretchr/testify/assert"
)

type Author struct {
	Name    string `json:"name"`
	Country string `json:"country"`
}

type Comment struct {
	NickName string    `json:"nickName"`
	Content  string    `json:"content"`
	Date     time.Time `json:"date"`
}

type Blog struct {
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	ViewCount uint64    `json:"viewCount"`
	Author    Author    `json:"author"`
	Date      time.Time `json:"date"`
	Comments  []Comment `json:"comments"`
}

func buildTestBlog() *Blog {
	t := time.Now()
	return &Blog{
		Title:     "Blog title1",
		Content:   "we are the blog content",
		ViewCount: 12,
		Author: Author{
			Name:    "robi",
			Country: "cn-ZH",
		},
		Date: t,
		Comments: []Comment{
			Comment{
				NickName: "lysu",
				Content:  "dajiangyou",
				Date:     t,
			},
		},
	}
}

func toString(data interface{}) string {
	d, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(d)
}

func TestPatchTopLevelString(t *testing.T) {
	assert := assert.New(t)
	p := patcher.Patcher{}
	b := buildTestBlog()
	ps := patcher.Patch{
		patcher.Path("title"): "hehe 233",
	}
	p.PatchIt(b, ps)
	assert.Equal("hehe 233", b.Title)
}
