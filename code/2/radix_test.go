package radix

import "testing"

const (
	Accept                  = "Accept"
	AcceptCharset           = "Accept-Charset"
	AcceptEncoding          = "Accept-Encoding"
	AcceptLanguage          = "Accept-Language"
	AcceptRanges            = "Accept-Ranges"
	Authorization           = "Authorization"
	CacheControl            = "Cache-Control"
	Cc                      = "Cc"
	Connection              = "Connection"
	ContentEncoding         = "Content-Encoding"
	ContentId               = "Content-Id"
	ContentLanguage         = "Content-Language"
	ContentLength           = "Content-Length"
	ContentRange            = "Content-Range"
	ContentTransferEncoding = "Content-Transfer-Encoding"
	ContentType             = "Content-Type"
	CookieHeader            = "Cookie"
	Date                    = "Date"
	DkimSignature           = "Dkim-Signature"
	Etag                    = "Etag"
	Expires                 = "Expires"
	Expect                  = "Expect"
	From                    = "From"
	Host                    = "Host"
	IfModifiedSince         = "If-Modified-Since"
	IfNoneMatch             = "If-None-Match"
	InReplyTo               = "In-Reply-To"
	LastModified            = "Last-Modified"
	Location                = "Location"
	MessageId               = "Message-Id"
	MimeVersion             = "Mime-Version"
	Pragma                  = "Pragma"
	Received                = "Received"
	Referer                 = "Referer"
	ReturnPath              = "Return-Path"
	ServerHeader            = "Server"
	SetCookieHeader         = "Set-Cookie"
	Subject                 = "Subject"
	TransferEncoding        = "Transfer-Encoding"
	To                      = "To"
	Trailer                 = "Trailer"
	UpgradeHeader           = "Upgrade"
	UserAgent               = "User-Agent"
	Via                     = "Via"
	XForwardedFor           = "X-Forwarded-For"
	XImforwards             = "X-Imforwards"
	XPoweredBy              = "X-Powered-By"
)

type (
	TreeTester struct {
		Tree
		logger func(format string, args ...interface{})
	}
)

//PrintTree: Print out current tree struct, it will using \t for tree level
func (t *TreeTester) PrintTree(currentNode *Node, treeLevel int) {
	if currentNode == nil {
		currentNode = &t.root
	}
	tabs := ""
	for i := 1; i < treeLevel; i++ {
		tabs = tabs + "\t"
	}

	if currentNode.isLeaf() {
		//Reach  the end point
		t.logger("%s[%d] Leaf key : %q value : %v\n", tabs, treeLevel, currentNode.leaf.key, currentNode.leaf.value)
		return
	}

	t.logger("%s[%d] Node has %d edges \n", tabs, treeLevel, len(currentNode.edges))
	for _, edge := range currentNode.edges {
		t.logger("%s[%d] NormalEdge [%q]\n", tabs, treeLevel, edge.label)
		t.PrintTree(edge.child, treeLevel+1)
	}

	if treeLevel == 1 {
		t.logger("Tree printed.\n\n")
	}
}

func TestLookup(t *testing.T) {
	rTree := &TreeTester{logger: t.Logf}
	rTree.Insert("test", 1)
	rTree.Insert("team", 2)
	rTree.Insert("trobot", 3)
	rTree.Insert("apple", 4)
	rTree.Insert("app", 5)
	rTree.Insert("tesla", 6)

	ret, find := rTree.Search("team")
	if !find || ret != 2 {
		t.Errorf("Lookup failed, expect '2', but get %v", ret)
	}

	ret, find = rTree.Search("apple")
	if !find || ret != 4 {
		t.Errorf("Lookup failed, expect '4', but get %v", ret)
	}

	ret, find = rTree.Search("tesla")
	if !find || ret != 6 {
		t.Errorf("Lookup failed, expect '6', but get %v", ret)
	}

	ret, find = rTree.Search("app")
	if !find || ret != 5 {
		t.Errorf("Lookup failed, expect '5', but get %v", ret)
	}

	rTree.Insert("app", 7)
	rTree.PrintTree(nil, 1)
	ret, find = rTree.Search("app")
	t.Log(ret, find)
	if !find || ret != 7 {
		t.Errorf("Insert update lookup failed, expect '7', but get %v", ret)
	}
}
