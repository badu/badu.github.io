---
title: My Thoughts On Net/Http Package - Week 1
tags: ["Golang", "Net", "Http", "Analysis"]
date: 2018-02-18
description : A deep dive into it net/http package. First glance.
---

## TL;DR

This series is about my questions and thoughts regarding net/http package. The process of learning is based on mistakes, therefor I'm inviting you to learn aside me.

You are allowed to judge the code. You are not allowed to judge the people.
 
## First Glance

I have a confession to make : despite the fact that ["Keep types close" rule](https://rakyll.org/style-packages/) is fair enough, the architect in me needs to reorganize the code in such a manner that a 10000 feet view to be possible. Creating a "types.go" file and storing there the structures, variables and constant declarations, allows me to see better the relationships between data (let's say models). 

I'm that kind of dude that prefers the "bottom up" approach about software development, so I can decide when a package grew too large or the separation of concerns is violated. One might say that creating "types.go" would violate the "no plurals" rule, but hey, "type" is a keyword in Go, isn't it?

My first action in this deep dive action was to break everything up inside the net/http package so I can have that distant look on it. One of the above mentioned rules states that "we organize code by their functional responsibilities". Using common sense with this rule and putting it besides the "types.go" would imply that all structs receiver would sit in their own file, wouldn't it? Yes, one might say that we'll have a huge collection of files, but rest assured : you will not have to navigate inside a 12000 lines of code file just to check a function's body. Also, another file that holds the "functional responsibilities" would be "utils.go", which hosts all the non receiver functions that compiler would inline or the package exposes publicly.

If a package has a large number of files and we need to keep some sort of track on who's who, we can apply even more splitting : let's say we have "types.go" which comes from both "server.go" and "client.go", but we don't want to mix those together. Seems to me a good idea to have "types_server.go" and "types_client.go" - easy to find, easy to read. Same applies to "utils.go". 

One could even create a "public.go" file which will host every function that his package exposes to the outside

If you want to have a look on what results after applying the technique described above, here is a [folder](https://github.com/badu/badu.github.io/blob/master/code/1) which contains the split files of [server.go](https://github.com/golang/go/blob/master/src/net/http/server.go).

#### First Note to Self

The first conclusion was regarding tests. I've written tests myself, regretfully in the same messy manner, but now I've made a note to self to never allow me to do so anymore. After all, tests are code too. Making tests hard to read would lead either to long conversations / documentation explaining the usage of your package in certain conditions, when seems easier to indicate a test that answers someone's question. By the nature of test files, they tend to be messy and ugly, because they need to cover scenarios that are uncommon, mostly malfunctioning conditions.

#### Preparing It

The work of splitting everything up took me a week and implied, besides what I've already told you:

* loads of TODO (and @comment tagged) comments leaving breadcrumbs for most of the altering operations or questions that remained unanswered. I've done this before without leaving traces so, this is note to self number two : never alter code without noting down the reasons.
* moving files from main package into a different package - the case of "client.go", the most obvious one. The initial reason for keeping them in the same package was probably the small size that it had.
* removed HTTP2 support for now, because it's a too big task to dive into that in the same time
* renaming imports and fixing private-public requirements (exposing some private functions so they can be accessed from another package)
* removing dead code or deprecations - which, of course, spoiled some tests (e.g. [TestClientTimeout](https://github.com/golang/go/blob/47f4e7a9768a613371ccd4a94a6b325fd603727e/src/net/http/client_test.go#L1168))

#### Internal Nettracer

Digging through the tests, I've found that some of them (like transport_test.go) needed "internal/nettrace" package. I quote "This package is purely internal for use by the net/http/httptrace package and has no stable API exposed to end users.".

While you would expect that this package to be truly internal, it seems that the "net" package is using it in production [lookup.go](https://github.com/golang/go/blob/e4bde0510465eecd4c8a8293418b1cbed1e0e623/src/net/lookup.go#L176) and [dial.go](https://github.com/golang/go/blob/424b0654f8e6c72f69e096f69009096de16e30fa/src/net/dial.go#L341). There is nothing wrong with that except the fact that if you are going to implement your own "nettrace" you just can't. Thus, tests that require it will fail : in transport_test.go : testTransportEventTrace, TestTransportMaxIdleConns and testTransportIDNA.

A solution at the moment would imply setting an interface in the context.Context, then recover that interface inside the lookup.go and dial.go code, thus decoupling the dependency of internal. I'm not sure it worth the effort, however, I'll try, at least to make an issue on github.

Because the above mentioned tests fail, I've totally removed them.

#### Cookies

For some reason - which might be syntactic sugar or just laziness of the users - even if both Response and Request structs have a Header field [Header map[string][]string](https://github.com/golang/go/blob/master/src/net/http/header.go#L19), they also expose Cookie struct by having these methods : Request [Cookie(name string) (*Cookie, error)](https://github.com/golang/go/blob/master/src/net/http/request.go#L373), 
[AddCookie(c *Cookie)](https://github.com/golang/go/blob/master/src/net/http/request.go#L384), [Cookies() []*Cookie](https://github.com/golang/go/blob/master/src/net/http/request.go#L362) and Response [Cookies() []*Cookie](https://github.com/golang/go/blob/master/src/net/http/response.go#L119).

Because I've decided to move all the client related code in it's own package (to be easier to read), I had to dump these methods and create functions with the same functionality (but not the same name, because Cookies() []*Cookie collision for both Response and Request).

To be [continued](/post/net-http-week2/).