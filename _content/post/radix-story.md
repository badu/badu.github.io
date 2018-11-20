---
title: A Radix Story
tags: ["Go", "Radix", "Tree", "Router", "Search"]
date: 2018-02-27
description : About routing and searching using radix trees.
---

## A Radix Story

#### Yes, All You Need is Standard "net/http"

There are, out there in the wild, an abundance of routers that a developer can use. In my opinion, one thing that they all share is the lack of separation of concerns. By separation of concerns I mean mixing handlers with routes and conditions, putting on top of that some cool optimized algorithm for looking up a route (or maybe I should say inspecting a string...). If you don't need everything those packages provide and you feel you should get rid of those training wheels, read on.

#### A Small Detour

I should probably write a separate article on the subject of what I consider to be a good Go developer, but here is short version of my opinion.

As you all can observe, the whole variety of packages are competing with each other on the matters of allocations and speed. From my point of view this hides a powerful truth about Go language and the developers who use it.

One of the main reasons for switching languages and becoming a Golang developer has to do with speed execution. For this reason and this reason only, I think that we can tell a good developer from a mediocre one, just by looking at his abilities to use the tools that he has and create new ones specialized for the problem that needs a solution.

Long story short, I strongly believe that a good Golang developer should have a wide enough knowhow and apply it whenever needed by adapting a generic solution to a particular case - for the sake of simplicity, speed and allocations.

#### What We Are Trying To Solve?

The problem is really simple : let's separate the concerns in such manner that we can adapt a sufficient small piece of code to our needs, whatever those needs might be.

In the case of a router, http.Request gives us access to the requested path. So, if we are able to process that path in a way that we need, we might not need a router at all. In the end we should be able to write a well anchored set of business logic rules without making the effort to adapt our mental model to a set of rules and conventions imposed by the authors of a package.

To give you a simple example, most routers won't allow you to have a `otherwise` route : match this route with this handler - and so on, but otherwise, put all other requests on this handler. This comes handy when you need to cover a dynamic defined route, like serving a request based on the slug of a title or the slug of category (title slug being the `otherwise` handler).

Let's try to forget about what a route really  is, by splitting it into three main properties : a http method (GET/POST/PUT/etc.), a path string and a function which needs to be called if that path string looks like something particular.

From now on, we're going to deal with that string only. Leaving the theory alone, the [applications](https://en.wikipedia.org/wiki/Radix_tree) of a radix tree are constructing a tree of pieces of strings which hold a reference to an object (fancy called associative arrays with keys).

Let's get coding:
```go
	package radix

	type (
		Tree struct {
			root Node
		}

		Edge struct {
			label  string
			parent *Node
			child  *Node
		}

		Edges []Edge

		leafNode struct {
			key   string
			value interface{}
		}

		Node struct {
			leaf  *leafNode
			edges Edges
		}
	)

	func (n *Node) isLeaf() bool {
		return n.leaf != nil && len(n.edges) == 0
	}

	func (n *Node) createNode(edgeKey, leafKey string, value interface{}) {
		leafNode := &Node{
			leaf: &leafNode{
				key:   leafKey,
				value: value,
			},
		}
		newEdge := Edge{
			label:  edgeKey,
			parent: n,
			child:  leafNode,
		}
		n.edges = append(n.edges, newEdge)
	}

	func (n *Node) createNodeWithEdges(newKey string, edgeKey string) *Node {
		if n.isLeaf() {
			//node is leaf node could not split, return nil
			return nil
		}

		for idx, edge := range n.edges {
			if edge.label == edgeKey {
				// backup for split
				oldNode := edge.child
				// create a new node
				newNode := &Node{}
				// replace current edge with a new one
				n.edges[idx] = Edge{
					label:  newKey,
					parent: n,
					child:  newNode,
				}
				// connect to original node
				remainKey := strings.TrimPrefix(edgeKey, newKey)
				newEdge := Edge{
					label:  remainKey,
					parent: newNode,
					child:  oldNode,
				}
				// append the edges with that new edge
				newNode.edges = append(newNode.edges, newEdge)
				return newNode
			}
		}
		return nil
	}

	// it's recursive
	func (t *Tree) insert(target *Node, edgeKey string, leafKey string, value interface{}) {
		// we've reached leaf
		if target.isLeaf() {
			if leafKey == target.leaf.key {
				// the same leaf key, update value
				// if overwriting values is by convention forbidden, should panic
				target.leaf.value = value
			} else {
				// insert leaf key value as new child node of target
				// original leaf node, became another leaf of target
				target.createNode(edgeKey, leafKey, value)
				// we have a convention here, regarding empty strings
				target.createNode("", target.leaf.key, target.leaf.value)
				target.leaf = nil
			}
			return
		}

		// second case, checking edges
		for _, edge := range target.edges {
			compare, found := longestPrefix(edgeKey, edge.label)
			if found {
				if compare == edge.label {
					// trim edge.label from new key
					nextTargetKey := strings.TrimPrefix(edgeKey, edge.label)
					// recurse
					t.insert(edge.child, nextTargetKey, leafKey, value)
				} else {
					// using compare to create new node and separate two edges
					splitNode := target.createNodeWithEdges(compare, edge.label)
					if splitNode == nil {
						panic("Unexpected error on creating new node and separating edges")
					}
					splitContainKey := strings.TrimPrefix(edgeKey, compare)
					splitNode.createNode(splitContainKey, leafKey, value)
				}
				return
			}
		}
		// new edge with new leaf key on leaf node
		target.createNode(edgeKey, leafKey, value)
	}

	func (t *Tree) Insert(what string, value interface{}) {
		//leaf key and edge key are the same
		t.insert(&t.root, what, what, value)
	}

	func (t *Tree) search(where *Node, what string) (interface{}, bool) {
		if where.isLeaf() {
			return where.leaf.value, true
		}
		for _, edge := range where.edges {
			if compare, found := longestPrefix(what, edge.label); found {
				nextSearchKey := strings.TrimPrefix(what, compare)
				return t.search(edge.child, nextSearchKey)
			}
		}
		return nil, false
	}

	func (t *Tree) Search(what string) (interface{}, bool) {
		return t.search(&t.root, what)
	}

	func longestPrefix(s1, s2 string) (string, bool) {
		found := false
		for i := 0; i < len(s1) && i < len(s2); i++ {
			if s1[i] != s2[i] {
				result := s1[:i]
				return result, found
			}
			found = true
		}

		if len(s1) > len(s2) {
			return s2, found
		} else if len(s1) == len(s2) {
			// special case : "" is not a subset of ""
			return s1, s1 == s2
		}

		return s1, found
	}
```

The code is pretty straight forward, however I've placed some comments here and there, just in case.

As you can see inside the [test](https://github.com/badu/badu.github.io/blob/master/code/2/radix_test.go), I've tested the search against some HTTP headers, just in case someone might find it useful to discard any unwanted headers inside a middleware. Later on, I'll probably benchmark the classical method of collecting HTTP headers against this method. However, this is not the point here.

Usually, we, developers need to match URLs like `/posts/{tag}/{id:[0-9]+}` in the HTTP path and while we're at it also extract those meaningful arguments. Things seem even more complicated when the string contains query values, like `/posts/{tag}?filter=date&author=3`.

#### The Star Lookup

Let's make the following convention : we don't care (at the moment) how an argument look like. For us, everything that is an argument is a star (`*`), leaving the decision of converting and validation to the next business logic level (which might be as well the handler itself).

Given the above convention, the path `/posts/{tag}/{id:[0-9]+}` would translate to `/posts/*/*`, ok?

Also, in order to capture the query values `/posts/{id:[0-9]+}?filter=date` would translate to `/posts/*?*`.

Here's the code :
```go
	func (e Edge) isStar() bool {
		return strings.HasPrefix(e.label, "*") || strings.HasPrefix(e.label, "/*") || strings.HasPrefix(e.label, "?*")
	}

	func (t *Tree) starSearch(where *Node, what string) (interface{}, bool) {
		if where.isLeaf() {
			return where.leaf.value, true
		}

		for _, edge := range where.edges {
			if edge.isStar() {
				// search key is empty and we're on the "/*" means we're looking for the last sibling of the edge
				if what == "" && strings.HasPrefix(edge.label, "/*") {
					continue
				}
				// different kind of star... ("*" or "?*")

				// remove the common parts
				remove, found := longestPrefix(what, edge.label)
				if found {
					what = strings.TrimPrefix(what, remove)
				}

				// split by slashes so we can build a new key
				parts := strings.Split(what, "/")

				switch len(parts) {
				case 1:
					// ok, we had one piece
					// looking for the question mark - might be handy to give up on this for speed
					index := strings.Index(what, "?")
					if index > 0 {
						//TODO : collect star key after question mark
						fmt.Println("Collect after question mark ", what[index:])
						// lookup question marks - down in the tree
						return t.starSearch(edge.child, what[index:])
					}
					// don't have a question mark, but we have a star (continue) - looking for the last sibling edge
					if strings.HasPrefix(edge.label, "?*") && remove != "?" {
						continue
					}
					//TODO : collect star key
					fmt.Println("Collect Path Part", what)
					// we have a star, no question mark - looking for the node leaf
					return t.starSearch(edge.child, "")
				default:
					//TODO : collect star key part of the path
					fmt.Println("Collect Path Part", parts[0])
					// building a new key with the parts that we have
					what = strings.Join(parts[1:], "/")
					return t.starSearch(edge.child, what)
				}
			}

			if compare, found := longestPrefix(what, edge.label); found {
				nextSearchKey := strings.TrimPrefix(what, compare)
				return t.starSearch(edge.child, nextSearchKey)
			}
		}

		return nil, false
	}

	func (t *Tree) StarSearch(what string) (interface{}, bool) {
		// TODO : use a mutex here if you are using this concurrent
		return t.starSearch(&t.root, what)
	}
```

Remember, you have to adapt this to your specific needs and for this reason I haven't imposed a way of capturing path parts that match star (`*`). I've left comments where you can do that.

In the above starSearch function there is a catch : if the edges slice is not sorted, so the last sibling is the star (`*`), the search will not work. To sort those slices, we need the following:
```go
	func (e Edges) Len() int {
		return len(e)
	}

	func (e Edges) Less(i, j int) bool {
		return e[i].label > e[j].label
	}

	func (e Edges) Swap(i, j int) {
		e[i], e[j] = e[j], e[i]
	}

	func (e Edges) Sort() {
		sort.Sort(e)
	}
```

The functions createNode and createNodeWithEdges will have the following additional piece of code:
```go
	if useSortEdges {
		// we're always sorting in reverse, so stars are last siblings
		newNode.edges.Sort()
	}
```

#### Recommendations

First, don't forget to use mutex for search - I've left a comment there - if you are using it in a concurrent way.

Secondly, the problem of defining one route is that you can't do this `/*/*` and expect it to work only in one route. In order to make it work you need an `exception string` (see TestStarOneRoute) and that exception string to look like `*/something-never-matched`.

Regarding leafNode struct value property, which is typed `interface{}` in the above code - you should do whatever you like with it. To give you an obvious example:
```go
	leafNode struct {
        key   string
		value http.HandlerFunc
	}
```

Last, but not least, you can alter things like adding a flag to the tree to mark it as dirty, then process all the edges in the tree and mark them as isStar `true/false` so `isStar() bool` function can be replaced like that. Or you could rewrite some parts to reduce allocations.

#### Other usages

Besides request routing, radix trees can be used in building ACLs (storing capabilities/permissions in the `value` property), spell checkers, auto-complete suggestions and [in memory database](https://github.com/hashicorp/go-memdb).

#### Code

The complete code (tests included) is [here](https://github.com/badu/badu.github.io/blob/master/code/2).
