package radix

import (
	"testing"
)

type (
	routeAndValue struct {
		r string
		v int
	}
	routes []routeAndValue
)

// PrintTree: Print out current tree struct, it will using \t for tree level
func (t *TreeTester) PrintStarTree(currentNode *Node, treeLevel int) {
	if currentNode == nil {
		currentNode = &t.root
	}
	tabs := ""
	for i := 1; i < treeLevel; i++ {
		tabs = tabs + "\t"
	}

	if currentNode.isLeaf() {
		// Reach  the end point
		t.logger("%s[%d] Leaf key : %q value : %v\n", tabs, treeLevel, currentNode.leaf.key, currentNode.leaf.value)
		return
	}

	t.logger("%s[%d] Node has %d edges \n", tabs, treeLevel, len(currentNode.edges))
	for _, edge := range currentNode.edges {
		if edge.isStar() {
			t.logger("%s[%d] StarEdge [%q]\n", tabs, treeLevel, edge.label)
		} else {
			t.logger("%s[%d] NormalEdge [%q]\n", tabs, treeLevel, edge.label)
		}
		t.PrintTree(edge.child, treeLevel+1)
	}

	if treeLevel == 1 {
		t.logger("Tree printed.\n\n")
	}
}

// This should work with only one route defined
func TestStarOneRoute(t *testing.T) {
	useSortEdges = true
	rTree := &TreeTester{logger: t.Logf}
	rTree.Insert("*/*", 5555)
	// without the below insertion, works fine, but you won't be able to collect the parameters correctly
	rTree.Insert("*/never-matched", 10000)
	ret, find := rTree.StarSearch("something/else")
	if !find || ret != 5555 {
		t.Errorf("Lookup failed, expect '5555', but got %v", ret)
	} else {
		t.Log("Ok `something/else` ", ret)
	}

	ret, find = rTree.StarSearch("do/else")
	if !find || ret != 5555 {
		t.Errorf("Lookup failed, expect '5555', but got %v", ret)
	} else {
		t.Log("Ok `do/else` ", ret)
	}

	rTree.PrintStarTree(nil, 1)
}

func TestStar(t *testing.T) {
	useSortEdges = true
	rTree := &TreeTester{logger: t.Logf}

	routes := routes{
		{
			"*", 11,
		}, {
			"test", 1,
		}, {
			"team", 2,
		}, {
			"trouble", 3,
		}, {
			"apple", 4,
		}, {
			"app", 5,
		}, {
			"app/blah", 555,
		}, {
			"app/blah/blah", 5555,
		}, {
			"app/blah/blah?*", 10000,
		}, {
			"tesla", 6,
		}, {
			"test/*", 12,
		}, {
			"test/*/*", 13,
		}, {
			"test/*/*?*", 14,
		}, {
			"tesla/copy/*?*", 202,
		}, {
			"tesla/particular/*?*", 220,
		}, {
			"tesla/*/*?*", 200,
		}, {
			"tesla/*/paste", 205,
		}, {
			"tesla/*/paste/*?*", 210,
		},
	}
	for _, route := range routes {
		rTree.Insert(route.r, route.v)
	}

	var ret interface{}
	var find bool

	ret, find = rTree.StarSearch("app/blah/blah")
	if !find || ret != 5555 {
		t.Errorf("Lookup failed, expect '5555', but got %v", ret)
	} else {
		t.Log("Ok `app/blah/blah` ", ret)
	}

	ret, find = rTree.StarSearch("app/blah/blah?filter=blah")
	if !find || ret != 10000 {
		t.Errorf("Lookup failed, expect '10000', but got %v", ret)
	} else {
		t.Log("Ok `app/blah/blah?filter=blah` ", ret)
	}

	ret, find = rTree.StarSearch("app/blah")
	if !find || ret != 555 {
		t.Errorf("Lookup failed, expect '555', but got %v", ret)
	} else {
		t.Log("Ok `app/blah` ", ret)
	}

	ret, find = rTree.StarSearch("tesla/copy/oops?search=blah")
	if !find || ret != 202 {
		t.Errorf("Lookup failed, expect '202', but got %v", ret)
	} else {
		t.Log("Ok `tesla/copy/oops?search=blah` ", ret)
	}

	ret, find = rTree.StarSearch("test/457/doo?search=string")
	if !find || ret != 14 {
		t.Errorf("Lookup failed, expect '14', but got %v", ret)
	} else {
		t.Log("Ok `test/457/doo?search=string` ", ret)
	}

	ret, find = rTree.StarSearch("test/123")
	if !find || ret != 12 {
		t.Errorf("Lookup failed, expect '12', but got %v", ret)
	} else {
		t.Log("OK `test/123` ", ret)
	}

	ret, find = rTree.StarSearch("test/123/abc")
	if !find || ret != 13 {
		t.Errorf("Lookup failed, expect '13', but got %v", ret)
	} else {
		t.Log("OK `test/123/abc` ", ret)
	}

	ret, find = rTree.StarSearch("tesla/457/paste/oops?search=blah")
	if !find || ret != 210 {
		t.Errorf("Lookup failed, expect '210', but got %v", ret)
	} else {
		t.Log("Ok `tesla/457/paste/oops?search=blah` ", ret)
	}

	ret, find = rTree.StarSearch("tesla/457/paste")
	if !find || ret != 205 {
		t.Errorf("Lookup failed, expect '205', but got %v", ret)
	} else {
		t.Log("Ok `tesla/457/paste` ", ret)
	}

	ret, find = rTree.StarSearch("tesla/457/doo?search=string")
	if !find || ret != 200 {
		t.Errorf("Lookup failed, expect '200', but got %v", ret)
	} else {
		t.Log("Ok `tesla/457/doo?search=string` ", ret)
	}

	ret, find = rTree.StarSearch("trouble")
	if !find || ret != 3 {
		t.Errorf("Lookup failed, expect '3', but got %v", ret)
	} else {
		t.Log("Ok `trouble` ", ret)
	}

	ret, find = rTree.StarSearch("something")
	if !find || ret != 11 {
		t.Errorf("Lookup failed, expect (universal) '11', but got %v", ret)
	} else {
		t.Log("Ok `something` ", ret)
	}

	rTree.PrintStarTree(nil, 1)
	t.Log("Test finished.")
}
