package radix

import (
	"fmt"
	"sort"
	"strings"
)

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
