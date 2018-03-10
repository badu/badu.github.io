---
title: Knowing when the world stops
tags: ["Golang", "Advanced", "Compiler", "Directives"]
date: 2018-03-05
description : I quote "This function is called with the world stopped, at the beginning of a garbage collection."
---

Last week, I took the time searching for patterns inside the main packages. Besides a bunch of `aha moments`, I've realized that some neat tricks can be used to achieve some goals otherwise achievable by applying different techniques.

For instance, let's say you have a pool or a cache. How do you call your cleanup function? 

Decisions regarding where to place that call can be made by testing and benchmarking. But what if there is another neat way to do so : just before the garbage collection runs, you can mount yourself a function and receive a call on it.

#### Compiler Directives

A [compiler directive](https://golang.org/cmd/compile/#hdr-Compiler_Directives) is a meta information that instructs the compiler on how to behave in certain conditions.

In our case, we're using the `sync.pool` [method](https://github.com/golang/go/blob/release-branch.go1.9/src/sync/pool.go#L218) of cleaning up the pool:

	```go
	//go:linkname runtime_registerPoolCleanup sync.runtime_registerPoolCleanup
	func runtime_registerPoolCleanup(cleanup func())
	```

Here's what we need to do :

 1. in our package, create a file named `empty.s`. As the name says, it's empty.
 2. in the file where we're going to declare the linkname directive, we have to import "unsafe" package. So, `import _ "unsafe"`.
 3. use the directive :
 
	```go
	//go:linkname registerCacheCleanupFn sync.runtime_registerPoolCleanup
	func registerCacheCleanupFn(f func())
	```
 4. in the same file, declare the `init()` function and call `registerCacheCleanupFn` with our cleaning function implementation as parameter.
 
That's it.

Advantages of using this technique are obvious. However, we have to keep in mind that our cleaning up implementation should NOT allocate and should NOT call any runtime functions - unless you think you are Mario (the plumber) and can deal with any kind of leak.

#### What else can we use?

If we to avoid importing "strings" just like [parse.go](https://github.com/golang/go/blob/release-branch.go1.9/src/net/parse.go#L86) inside "net" package does, but still be able to call IndexByte, we need the following declaration : 

```go
//go:linkname ByteIndex strings.IndexByte
func ByteIndex(s string, c byte) int
```

You might find these below useful as well.

Bytes :

```go
//go:linkname BytesEqual bytes.Equal
func BytesEqual(x, y []byte) bool
```
```go
//go:linkname IndexByte bytes.IndexByte
func IndexByte(s []byte, c byte) int
```
```go
//go:linkname Compare bytes.Compare
func Compare(a, b []byte) int
```

Time :

```go
//go:linkname TimeSleep time.Sleep
func TimeSleep(ns int64)
```
```go
//go:linkname StartTimer time.startTimer
func StartTimer(t *timer)
```
```go
//go:linkname StopTimer time.stopTimer
func StopTimer(t *timer) bool
```
```go
//go:linkname PollNano internal/poll.runtimeNano
func PollNano() int64
```
```go
//go:linkname TimeNano time.runtimeNano
func TimeNano() int64
```
```go
//go:linkname Now time.now
func Now() (sec int64, nsec int32, mono int64) 
```

#### A Side Note

By the way : I've noticed that strings.Index is sometimes unnecessary called when the second parameter is just a byte (e.g. `strings.Index(url, "?")` in [server.go](https://github.com/golang/go/blob/release-branch.go1.9/src/net/http/server.go#L2002)). I know it is a micro optimization, but hey, let's use the right tool for the right job, shall we?

Same observation goes for bytes.IndexByte too.

Another observation regards strings.TrimPrefix and strings.TrimSuffix - I've notice that sometimes developers checks if string has prefix or suffix before calling them, which is unnecessary because the check is done inside those functions. Perhaps it would be better to change the signatures of those functions like this :

```go
func TrimPrefix(s, prefix string) (bool, string) // returning true if string had that prefix and the trimmed string
```

#### Directive Wish

Would be really useful for testing to have a directive that instructs the compiler to include or exclude portions of code, thus we can avoid including testing portions in the production.