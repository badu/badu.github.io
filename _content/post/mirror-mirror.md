---
title: Mirror Mirror on ...
tags: ["Golang", "Advanced", "Reflect", "Unsafe"]
date: 2018-03-10
description : Of Mice (Unsafe) and Men (Reflect)
---

#### TL;DR

While I was mentoring, I encouraged my pupils to break things so they understand how they work. Using reflect package seems easy, but understanding the mechanics is hard. So, this week, following my own advice, I've tried to create my own reflect package. Here is what I've learned.

#### Playing with Fire

Most of the articles on the subject I've read have (more or less) the following advice : "if you find yourself doing this in a real program, stop immediately and seek help. You are doing something wrong. You’ve been warned!". Now, wait a minute, mister. That is hypocrisy!

If you take a look at the importers of [reflect](https://godoc.org/reflect?importers) you will easily find that using the `fmt` implies you are using reflect. Using "unsafe" features in Golang is only for developers that develop the language itself? Maybe looking on importers of [unsafe](https://godoc.org/unsafe?importers) tells you otherwise.

#### `unsafe` - after that we `reflect`

I quote from the documentation : "unsafe.Pointer type  allows a program to defeat the type system and read and write arbitrary memory. It should be used with extreme care".

Let's say you have type `John` which you are trying to convert to type `Ivan`. The documentation states that Ivan has to be smaller or equal with John (in terms of properties it has) and those to share the `equivalent memory layout.

Let's code:

```go
func TestIsJohnIvan(t *testing.T) {
	type John struct {
		Name   string
		Age    uint
		Powers uint
	}

	type Ivan struct {
		givenName  string // yes, you can use private fields
		_          uint
		ThirdField uint
	}
	
	john := John{Name: "John", Age: 40, Powers: 3}

	ivan := *(*Ivan)(unsafe.Pointer(&john))
	t.Logf("John as Ivan : GivenName %v ThirdField %d", ivan.givenName, ivan.ThirdField)
}
``` 

First observation is that you can violate access to private fields using this conversion. Secondly, as long as you respect the same number of fields and their types, you can omit properties. You can violate the second rule and get unexpected results, as below:

```go
	type ShortIvan struct {
		Age uint
	}
	smallIvan := *(*ShortIvan)(unsafe.Pointer(&john))
	t.Logf("Small Ivan (just powers) : %v", smallIvan.Age)
```

You would expect that age to be 40, but it's not : it's 5717318. Why? Because an uint is built by taking the required value from the Name property of John. The correct way to get a smaller Ivan is to omit the name property (observe that the third property is omitted too):

```go
	type ShortIvan struct {
		_ string
		Age uint
	}
```

What if you violate the first rule, which states that types have to have an equal amount of properties:

```go
	type UpgradedIvan struct {
		//_ string // adding this at the beginning crashes
		Name    string
		Age     uint
		Powers  uint
		Address string // will get filled with the Name 
		Guns    uint   // will get filled with Age 
		//Say     string // adding yet another one will crash : "bad pointer in frame"
		//Data []byte // same adding this or more
		AFloat float32 // adding a different type seems safe
	}
	chuckNorris := *(*UpgradedIvan)(unsafe.Pointer(&john))
	t.Logf("Chuck Ivan : %v", chuckNorris)
```

Well, it works, but with side effects : Address gets filled with same value as Name, Guns with Age and AFloat get a value of 4e-45. So, this the non-safety point that a developer should never touch. As long as we're respecting the rules, it's safe to play unsafe.

Also, `upgrading` John seems better (simpler) by using embedding:

```go
	type EmbeddedJohn struct {
		John
		Address string // will get filled with the Name ??? Weird huh
		Guns    uint   // will get filled with Age ???
	}
	// convert John to EmbeddedJohn
	upgradedIvan := EmbeddedJohn{John: john}
	t.Logf("Upgraded Ivan : %v", upgradedIvan)
```

Surely, the bellow code is dangerous if it is misused. The code speaks for itself:

```go
func TestAlteredPeople(t *testing.T) {
	type John struct {
		Name    string
		Age     int
		Altered bool
	}

	john := John{Name: "John", Age: 30, Altered: false}

	ptrToJohn := unsafe.Pointer(&john)
	ptrToName := (*string)(unsafe.Pointer(uintptr(ptrToJohn) + unsafe.Offsetof(john.Name)))
	ptrToAge := (*int)(unsafe.Pointer(uintptr(ptrToJohn) + unsafe.Offsetof(john.Age)))
	ptrToAltered := (*bool)(unsafe.Pointer(uintptr(ptrToJohn) + unsafe.Offsetof(john.Altered)))

	*ptrToName = "Chuck"
	*ptrToAge = 100000
	*ptrToAltered = true

	t.Logf("Now John is %v", john)
}
```

#### `Unsafe` conclusions

The unsafe package is serving for Go compiler instead of Go runtime, because it has facilities for low-level programming including operations that violate the type system.

I would never use the above method of conversion, but investigation was needed because of what's about to be described regarding reflect. 

```go
type Point struct {
	x, y int
}

func Extract(ptr unsafe.Pointer, size uintptr) []byte {
	out := make([]byte, size)
	for i := range out {
		out[i] = *((*byte)(unsafe.Pointer(uintptr(ptr) + uintptr(i))))
	}
	return out
}

func TestExtract(t *testing.T) {
	p := Point{3, 4}
	mem := Extract(unsafe.Pointer(&p), unsafe.Sizeof(p))
	t.Logf("What's the Point? %v", mem)
}
```

Yes, you can extract the content of the memory, but what's the point? Well, a friend of mine (with the same name and the same passion for Golang) might find this as a useful way to hide sensitive data, by reversing the extract into carefully  filling it with secrets which comes from somewhere else.

#### `reflect`

If you kept in mind that `unsafe` is about the compiler and not the runtime, here is the proof:

```go
func TestInTheBeginning(t *testing.T) {
	type r struct {
		sz  uintptr
		dt  uintptr
		_   uint32
		f   uint8
		_   uint8
		_   uint8
		knd uint8
		_   *struct{}
		c   *byte
		str int32
		w   int32
	}
	type e struct {
		abracadabra *r
	}
	t := func(p interface{}) *r {
		return (*(*e)(unsafe.Pointer(&p))).abracadabra
	}
	p := Point{3, 4}
	v := t(&p)
	t.Logf("After looking in the mirror : %v %v %v %v %v %v", v.sz, v.dt, v.f, v.knd, v.str, v.w)
}
```

Once you run the above test, you will get the properties filled in with some values which seem pure magic. But there must be an explanation. 
We didn't import `reflect` package. Also, the code is unreadable thus proving there is no magic convention like structs named in certain way or properties have some particular names. 

So, what happen? Well, these data structures (`e` and `r` types) are [known](https://github.com/golang/go/blob/master/src/cmd/compile/internal/gc/reflect.go) to the compiler which does it's job and at the runtime we're getting those results. To reinforce that truth, if we're replacing that `t` function with it's body `v :=(*(*e)(unsafe.Pointer(&p))).abracadabra`, it won't work anymore. And even more, if we're changing the parameter type of the `t` function from `interface{}` to `*Point` it will not work as expected.

If you [look](https://github.com/golang/go/blob/master/src/reflect/type.go#L297) in reflect package, you will see that `rtype` struct looks exactly the same as our `r` struct, even if the properties are named different. Same goes for [`emptyInterface`](https://github.com/golang/go/blob/master/src/reflect/value.go#L181) and our `e` struct - despite the fact that we are not using the `word` property - remember omitting properties in the unsafe example above?

#### Building your own reflection package

Can you build your own `reflect` package? So far my conclusion is yes, you can. At least for reading and writing the properties of structs it's quite easy.

However, I've encountered some problems that I want to present here. First, the (long but minimal) code (mostly copy pasted from reflect):
```go

import (
	"testing"
	"unsafe" // also for linkname
)
const (
	Invalid       Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	Array
	Chan
	Func
	Interface
	Map
	Ptr
	Slice
	String
	Struct
	UnsafePointer
)
// minimal version of flags
const (
	tflagUncommon  tflag = 1 << 0
	tflagExtraStar tflag = 1 << 1
)

// minimal version of kinds
const (
	kindMask = (1 << 5) - 1
)

// minimal version of types
type (
	Kind uint
	nameOff int32
	typeOff int32
	textOff int32
	tflag uint8
	name struct {
		bytes *byte
	}
	uncommonType struct {
		pkgPath nameOff
		mcount  uint16
		_       uint16
		moff    uint32
		_       uint32
	}
	rtype struct {
		size       uintptr
		ptrdata    uintptr
		hash       uint32
		tflag      tflag
		align      uint8
		fieldAlign uint8
		kind       uint8
		alg        *typeAlg
		gcdata     *byte
		str        nameOff
		ptrToThis  typeOff
	}
	typeAlg struct {
		hash  func(unsafe.Pointer, uintptr) uintptr
		equal func(unsafe.Pointer, unsafe.Pointer) bool
	}
	method struct {
		name nameOff
		mtyp typeOff
		ifn  textOff
		tfn  textOff
	}
	structField struct {
		name       name
		typ        *rtype
		offsetAnon uintptr
	}
	structType struct {
		rtype `reflect:"struct"`
		pkgPath name
		fields  []structField
	}
	emptyInterface struct {
		typ  *rtype
		word unsafe.Pointer
	}
	stringHeader struct {
		Data unsafe.Pointer
		Len  int
	}
)
// utility function to find out the offsets
func add(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}

func (n name) isExported() bool {
	return (*n.bytes)&(1<<0) != 0
}

func (n name) name() (s string) {
	if n.bytes == nil {
		panic("no bytes present.")
	}
	b := (*[4]byte)(unsafe.Pointer(n.bytes))

	hdr := (*stringHeader)(unsafe.Pointer(&s))
	hdr.Data = unsafe.Pointer(&b[3])
	hdr.Len = int(b[1])<<8 | int(b[2])
	return s
}

func (t *rtype) nameOff(off nameOff) name {	
	return name{(*byte)(resolveNameOff(unsafe.Pointer(t), int32(off)))} 
}

func (t *rtype) typeOff(off typeOff) *rtype { 
	return (*rtype)(resolveTypeOff(unsafe.Pointer(t), int32(off))) 
}

func (t *rtype) Kind() Kind { return Kind(t.kind & kindMask) }

func (t *rtype) String() string {
	s := t.nameOff(t.str).name()
	if t.tflag&tflagExtraStar != 0 {
		return s[1:]
	}
	return s
}

func (t *uncommonType) methods() []method {
	if t.mcount == 0 {
		panic("zero methods.")
	}
	return (*[1 << 16]method)(add(unsafe.Pointer(t), uintptr(t.moff)))[:t.mcount:t.mcount]
}

func (t *rtype) uncommon() *uncommonType {
	if t.tflag&tflagUncommon == 0 {
		panic("bad flag.")
	}
	if t.Kind() != Struct {
		panic("not struct.")
	}
	type u struct {
		structType
		u uncommonType
	}
	ptrToT := unsafe.Pointer(t)
	return &(*u)(ptrToT).u
}

func (t *rtype) exportedMethods() []method {
	ut := t.uncommon()
	if ut == nil {
		panic("nil uncommon.")
	}
	allMethods := ut.methods()
	allExported := true
	for _, method := range allMethods {
		name := t.nameOff(method.name)
		if !name.isExported() {
			allExported = false
			break
		}
	}
	var methods []method
	if allExported {
		methods = allMethods
	} else {
		methods = make([]method, 0, len(allMethods))
		for _, m := range allMethods {
			name := t.nameOff(m.name)
			if name.isExported() {
				methods = append(methods, m)
			}
		}
		methods = methods[:len(methods):len(methods)]
	}
	return methods
}
// short variant of the TypeOf in the reflect package.
func TypeOf(i interface{}) *rtype {
	return (*(*emptyInterface)(ptrToI)).typ
}
```

Of course, to run tests, we have to create an `empty.s` file in the same folder and to add the linkname directives for two functions:
```go
//go:linkname resolveTypeOff reflect.resolveTypeOff
func resolveTypeOff(rtype unsafe.Pointer, off int32) unsafe.Pointer

//go:linkname resolveNameOff reflect.resolveNameOff
func resolveNameOff(ptrInModule unsafe.Pointer, off int32) unsafe.Pointer
```

On the `Point` struct declared above, we're adding the followings:

```go
func (p Point) AnotherMethod(scale int) int {
	return -1
}
func (p Point) Dist(scale int) int {
	return p.x*p.x*scale + p.y*p.y*scale
}
func (p Point) NoArgs() {
	println("NoArgs called.")
}
func (p Point) TotalDist(points ...Point) int {
	tot := 0
	for _, q := range points {
		dx := q.x - p.x
		dy := q.y - p.y
		tot += dx*dx + dy*dy
	}
	return tot
}
func (p Point) NoArgsButReturn() string {
	return "something"
}
```

And finally, the test :
```go

func TestMethod(t *testing.T) {
	p := Point{3, 4}
	pType := TypeOf(p)
	t.Logf("%v", pType)
	methods := pType.exportedMethods()
	for idx, method := range methods {
		name := pType.nameOff(method.name)
		typ := pType.typeOff(method.mtyp)
		t.Logf("%d : Method %q %v %v %v\n", idx, name.name(), typ, method.tfn, method.ifn)
	}
}
```

When we run this test, we're going to see that the methods signature are reported differently than what we've declared. This means we are not doing something that `reflect` package does. 

Our version of TypeOf function doesn't return an interface and also, that interface is built by calling toPtr() method of the rType. However, with that code added, the problem still doesn't get fixed.

Adding the following code, fixes the test (the signatures are correct).

```go
	reflect.TypeOf(p).Method(0)
```

Seems the function `func addReflectOff(ptr unsafe.Pointer) int32` which is implemented in the runtime package gets called from `reflect` package which creates [reflectOffs structs](https://github.com/golang/go/blob/34db5f0c4d80b8fe3fb4b5be90efd9ee92bd1d4d/src/runtime/type.go#L147) for later lookups.

In the larger version (my own version of reflect), all Value.Call() tests were failing in a segmentation fault, with no apparent reason - the code being the same as in reflect package. For this reason I've presented you with this small test and it's conclusions.

#### Conclusion

It took me four days to learn the internals and modify the reflect package for my needs, but in the end I've done it and later I will probably integrate it into the [reflector](https://github.com/badu/reflector) package.

Probably the lack of documentation made things harder to understand and follow. Probably some things are never meant to be - that - public, due to some sort of programming language politics. Who knows but mostly who cares?

I encourage you to take my advice and break things so you can learn how they work, how other developers solved problems that you cannot think about while just reading the code.

