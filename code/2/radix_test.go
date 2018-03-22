package radix

import (
	"testing"
)

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

var (
	headers = []string{
		Accept,
		AcceptCharset,
		AcceptEncoding,
		AcceptLanguage,
		AcceptRanges,
		Authorization,
		CacheControl,
		Cc,
		Connection,
		ContentEncoding,
		ContentId,
		ContentLanguage,
		ContentLength,
		ContentRange,
		ContentTransferEncoding,
		ContentType,
		CookieHeader,
		Date,
		DkimSignature,
		Etag,
		Expires,
		Expect,
		From,
		Host,
		IfModifiedSince,
		IfNoneMatch,
		InReplyTo,
		LastModified,
		Location,
		MessageId,
		MimeVersion,
		Pragma,
		Received,
		Referer,
		ReturnPath,
		ServerHeader,
		SetCookieHeader,
		Subject,
		TransferEncoding,
		To,
		Trailer,
		UpgradeHeader,
		UserAgent,
		Via,
		XForwardedFor,
		XImforwards,
		XPoweredBy,
	}
)

type (
	TreeTester struct {
		Tree
		logger func(format string, args ...interface{})
	}
)

// PrintTree: Print out current tree struct, it will using \t for tree level
func (t *TreeTester) PrintTree(currentNode *Node, treeLevel int) {
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
		t.logger("%s[%d] NormalEdge [%q]\n", tabs, treeLevel, edge.label)
		t.PrintTree(edge.child, treeLevel+1)
	}

	if treeLevel == 1 {
		t.logger("Tree printed.\n\n")
	}
}

func prepareTest(t *testing.T, tree *TreeTester) {
	for index, header := range headers {
		tree.Insert(header, index)
	}
	t.Log("Test prepared.")
}

func TestLookup(t *testing.T) {
	rTree := &TreeTester{logger: t.Logf}
	prepareTest(t, rTree)

	ret, find := rTree.Search(Accept)
	if !find {
		t.Error("Lookup failed")
	} else {
		t.Logf("Found : %v", ret)
	}

	ret, find = rTree.Search(AcceptLanguage)
	if !find {
		t.Error("Lookup failed")
	} else {
		t.Logf("Found : %v", ret)
	}

	rTree.Insert("Foo-Header", 7)
	ret, find = rTree.Search("Foo-Header")
	if !find || ret != 7 {
		t.Errorf("Insert update lookup failed, expect '7', but get %v", ret)
	}
	t.Log(find, " found freshly inserted ", ret)

	rTree.PrintTree(nil, 1)

}
