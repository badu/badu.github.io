---
title: A Radix Story 
tags: ["Golang", "Radix", "Tree", "Router", "Search"]
date: 2018-02-26
description : About routing and searching using radix trees.
---

## A Radix Story

#### Yes, All You Need is Standard "net/http"

There are, out there in the wild, an abundance of routers that a developer can use. One thing that they all share is the lack of separation of concerns. By separation of concerns I mean mixing handlers with routes and conditions, putting on top of that some cool optimized algorithm for looking up a route (or maybe I should say inspecting a string...).

#### A Small Detour

I should probably write a separate article on the subject of what I consider to be a good Go developer, but here is short version of my opinion.

As you all can observe, the whole variety of packages are competing with each other on the matters of allocations and speed. From my point of view this hides a powerful truth about Go language and the developers who use it.

One of the main reasons for being a Golang developer has to do with speed execution. For this reason and this reason only, I think that we can tell a good developer from a mediocre one, just by looking at his abilities to use the tools that he has and create new ones specialized for the problem that needs a solution.

Long story short, I strongly believe that a good Golang developer should have a wide enough knowhow and apply it whenever needed by adapting a generic solution to a particular case - for the sake of simplicity, speed and allocations.

#### What We Are Trying To Solve?

The problem is really simple : let's separate the concerns in such manner that we can adapt a sufficient small piece of code to our needs, whatever those needs might be.

In the case of a router, http.Request gives us access to the requested path. So, if we are able to process that path in a way that we need, we might not need a router at all, because we would be able to write one ourselves well anchored in the business logic that we need. 

Let's try to forget about what really a route is, by splitting it into three properties of it : a http method (GET/POST/PUT/etc.), a string which is the actual route itself and a function which needs to be called if that string looks like something particular.

For now, we're going to deal with that string only. Leaving the theory alone, the [applications](https://en.wikipedia.org/wiki/Radix_tree) of a radix tree are constructing a tree of pieces of strings which hold a reference to an object.

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

The complete code (tests included) is [here](https://github.com/badu/badu.github.io/blob/master/code/2).

I'm sure the code is pretty straight forward, however I've placed some comments here and there.

As you can see inside the [test](https://github.com/badu/badu.github.io/blob/master/code/2/radix_test.go), I've tested the search against some HTTP headers, just in case someone might find it useful to discard any unwanted headers inside a middleware.

Usually, we, developers need to match URLs like "/posts/{tag}/{id:[0-9]+}" in our router and while we're at it also extract those arguments. Things seem even more complicated when the string contains query values, like this "/posts/{tag}?filter=date&author=3".

#### The Star Lookup

Let's make the following convention : we don't care (for the moment) how the arguments look like. For us, everything that is an argument is a star, leaving the decision of converting and validation to the next business logic level (which might be as well the handler itself).

[...] The logic is to find the route, then to check the method and then to pass it to the business logic that deals with that.

[...] A solution would be to split your routing business logic into small, independent chunks and express them in their own handlers.