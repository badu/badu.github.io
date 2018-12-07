---
title: Interview Questions for Go Developer Position - Part II
tags: ["Go", "Developer", "Interview"]
date: 2018-12-07
description: Measuring And Classifying Go Developer Knowledge
---

#### Foreword

See [part 1](/post/interviewing-go-developer-part-1/).

Q: What will return the following code?

```go
func returnNormal() int {
	i := 1
	defer func() { i++ }()
	return i
}

func returnNamed() (i int) {
	i = 1
	defer func() { i++ }()
	return i
}

func main() {
	fmt.Printf("returnNormal() = %d\n", returnNormal())
	fmt.Printf("returnNamed() = %d\n", returnNamed())
}
```

A: `returnNormal() = 1` and `returnNamed() = 2` [3 points] (why : lazy evaluation)

---

Q: What will print the following code?

```go
type Inner struct {}

type InnerAgain struct {}

type A struct {
    Inner
    InnerAgain
    FieldA string
}

func (i Inner) String() string {
    return "anything"
}

func (i InnerAgain) String() string {
    return "nothing"
}

func main() {
    myA := A{FieldA: "A"}
    fmt.Printf("%v", myA)
}
```

A: `{anything nothing A}` [3 points] (why : ambiguous which String() to invoke)

---

Q: Can I convert a []T to an []interface{}?

A: Code below:

```go
t := []int{1, 2, 3, 4}
s := make([]interface{}, len(t))
for i, v := range t {
    s[i] = v
}
```

---

Q: Can I convert []T1 to []T2 if T1 and T2 have the same underlying type?

A: Code below:

```go
type T1 int
type T2 int
var t1 T1
var x = T2(t1) // OK
var st1 []T1
```

---

Q: What's wrong with the following code?

```go
type info struct{
	data string
}

test := []info{{data:"data 1"}, {data:"data 2"}, {data:"data 3"}}
list := make([]*string, 0)
for _, k := range test{
    list = append(list, &k.data)
}
for _, s := range list {
    fmt.Println(*s)
}
```

A: Don't use pointers inside loops [1 point]
A: Here. I've fixed it [3 points]

```go
type info struct{
	data string
}

test := []info{{data:"data 1"}, {data:"data 2"}, {data:"data 3"}}
list := make([]*string, 0)
for _, k := range test{
	nk := k // the pointer
    list = append(list, &nk.data)
}
for _, s := range list {
    fmt.Println(*s)
}
```

---

Q: What will print the following code?

```go
type People struct{}

func (p *People) ShowA() {
	fmt.Println("showA")
	p.ShowB()
}

func (p *People) ShowB() {
	fmt.Println("showB")
}

type Teacher struct {
	People
}

func (t *Teacher) ShowB() {
	fmt.Println("teacher showB")
}

func main() {
	t := Teacher{}
	t.ShowA()
}
```

A: `showA`
`showB` [3 points]

---

Q: Does this compile ?

```go
func main() {
	i := GetValue()

	switch i.(type) {
	case int:
		println("int")
	case string:
		println("string")
	case interface{}:
		println("interface")
	default:
		println("unknown")
	}

}

func GetValue() int {
	return 1
}

```
A: No [1 point]
A: No, because `i` is not an interface, so you cannot `type switch` on it [3 points]

---

Q: What will produce the following code ?

```go

type Param map[string]interface{}

type Show struct {
	Param
}

func main() {
	s := new(Show)
	s.Param["AValue"] = 10000
}
``` 

A : panic: assignment to entry in nil map [1 point]


---

Q: Does this compile ? 

```go
type student struct {
	Name string
}

func printName(v interface{}) {
	switch msg := v.(type) {
	case *student, student:
		fmt.Println(msg.Name)
	}
}
```

A: No [1 point]
A: No, error will be `type interface {} is interface with no methods` [3 points]

--- 

Q: Is there a problem with the code below ?

```go

type People struct {
	Name string
}

func (p *People) String() string {
	return fmt.Sprintf("print: %v", p)
}

func main() {
	p := &People{}
	p.String()
}
```

A: Yes [1 point]
A: Yes, will recurse infinitely [3 points]

---

To be continued.
