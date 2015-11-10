# go-el

Expression language(EL) to manipulate Golang structure data. Its main purpose is to find `reflect.Value` by Expression, then do some reading and writing.

# Installation

Simple as it takes to type the following command:

    go get github.com/lysu/go-el

and import with

    import github.com/lysu/go-el    

# Usage

## Example Data

As example, we have some data like this:

    type Comment struct {
    	NickName string
    	Content  string
    	Date     time.Time
    }

    type Author struct {
      Name string
    }

    type Blog struct {
    	Title      string
    	RoleState  map[string]uint
    	CommentIds []uint64
    	Comments   map[string]*Comment
    }

    func (b Blog) FirstComment() *Comment {
    	return b.Comments["0"]
    }

then we init them with some test data:

    b := &Blog{
      Title:      "Blog title1",
      CommentIds: []uint64{1, 3},
      Comments: map[string]*Comment{
        "0": {
          NickName: "000",
          Content:  "test",
          Date:     time.Now(),
        },
        "1": {
          NickName: "u1",
          Content:  "test",
          Date:     time.Now(),
        },
        "3": {
          NickName: "tester",
          Content:  "test hehe...",
          Date:     time.Now(),
        },
      },
      Author: Author{
          Name: "Author 1",
      },
      RoleState: map[string]uint{},
    }

## Expression

Using `el.Expression`, we can navigate from root(`b`) to anywhere in this structure.

#### 1. To field

    exp := el.Expression("Title")
    v, _ := exp.Execute(&data)
    fmt.Printf("%v\n", v.interface()) //==> Blog title1

#### 2. To nested field

    exp := el.Expression("Author.Name")
    v, _ := exp.Execute(&data)
    fmt.Printf("%v\n", v.interface()) //==> Author 1

#### 3. To slice/array/string item

    exp := el.Expression("CommentIds[0]")
    v, _ := exp.Execute(&data)
    fmt.Printf("%v\n", v.interface()) //==> 1

#### 4. To map item

    exp := el.Expression("Comments["3"].NickName")
    v, _ := exp.Execute(&data)
    fmt.Printf("%v\n", v.interface()) //==> tester

#### 5. Item in`[]` also can be another Expression

    exp := el.Expression("Comments["CommentIds[0]].NickName")
    v, _ := exp.Execute(&data)
    fmt.Printf("%v\n", v.interface()) //==> u1

#### 6. Call function

function can return only `ONE` result

    exp := el.Expression("FirstComment().Content")
    v, _ := exp.Execute(&data)
    fmt.Printf("%v\n", v.interface()) //==> test  

#### 7. Modify Value

After `Execute` expression, we got a `relfect.Value`, we also can use it to modify data, e.g.


    exp := el.Expression("FirstComment().Content")
    v, _ := exp.Execute(&data)
    v.SetString("1111")

will let first comment with value `1111`

Beside that we recommend users take a moment to look [The Laws of Reflection](http://blog.golang.org/laws-of-reflection), take care some limition that reflect has.   

## Patcher

Base on Expression, we also provide a tool named `Patcher`, the purpose of it is to let use modify object with expression easier and be batched.

We found it's very useful to build HTTP [Patch](http://tools.ietf.org/html/rfc5789) API to partial update entity


    ps := p.Patch{
      "Author.Name":                      "ほん",
      "Comments[CommentIds[0]].NickName": "私",
      "roleState[100]":                   uint(100),
    }
    err := patcher.PatchIt(b, ps)

This will modify three properties at once~ (but we still meet some rule of refect, like map-value use ptr.. and so on)    

## More

See our Example in Unit-Test:

- [expression](https://github.com/lysu/go-el/blob/master/expression_test.go)  
- [patcher](https://github.com/lysu/go-el/blob/master/patcher_test.go)

## TODO

generate expression between two data..like diff..- -?

# Thanks

- Many code was extract from [flosch/pongo2](https://github.com/flosch/pongo2) --- An cool template-engine
- The presentation by Rob Pike titled [Lexical Scanning in Go](http://cuddle.googlecode.com/hg/talk/lex.html#landing-slide)
