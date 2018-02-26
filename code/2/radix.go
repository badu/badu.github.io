package radix

import "strings"

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

var (
	useSortEdges bool
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
	if useSortEdges {
		// we're always sorting in reverse, so stars are last siblings
		n.edges.Sort()
	}
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
			if useSortEdges {
				// we're always sorting in reverse, so stars are last siblings
				n.edges.Sort()
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
			if useSortEdges {
				// we're always sorting in reverse, so stars are last siblings
				newNode.edges.Sort()
			}
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
