package themeapi

import (
	"bytes"
	"sort"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const quote = `"`

// escaper is the one escape rule of text and attribute values.
var escaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&quot;",
)

// escape rewrites s with escaper.
func escape(s string) string {
	return escaper.Replace(s)
}

// Canonical parses an HTML fragment as the browser does in a body and
// returns its canonical DOM form, the form of the frozen DOM goldens. It keeps
// what a theme can observe: the elements, their order, their attributes but
// id, and every text node, whitespace included. It drops what a theme cannot
// observe: the attribute order (sorted by name), the id attributes and the
// serialization (one escape rule for text and attribute values). The
// serializer is Umono's own, so a golang.org/x/net/html update does not change
// the goldens.
func Canonical(fragment []byte) (string, error) {
	nodes, err := html.ParseFragment(bytes.NewReader(fragment), &html.Node{Type: html.ElementNode, Data: "body", DataAtom: atom.Body})
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, n := range nodes {
		writeNode(&b, n)
	}
	return b.String(), nil
}

// writeNode writes the canonical form of n and its descendants to b.
func writeNode(b *strings.Builder, n *html.Node) {
	switch n.Type {
	case html.TextNode:
		b.WriteString(escape(n.Data))
	case html.CommentNode:
		b.WriteString("<!--")
		b.WriteString(n.Data)
		b.WriteString("-->")
	case html.ElementNode:
		b.WriteString("<")
		b.WriteString(n.Data)
		attrs := make([]html.Attribute, len(n.Attr))
		copy(attrs, n.Attr)
		sort.Slice(attrs, func(i, j int) bool {
			return attrs[i].Key < attrs[j].Key
		})
		for _, a := range attrs {
			if a.Key == "id" {
				continue
			}
			b.WriteString(" ")
			b.WriteString(a.Key)
			b.WriteString("=")
			b.WriteString(quote)
			b.WriteString(escape(a.Val))
			b.WriteString(quote)
		}
		b.WriteString(">")
		if voidElements[n.Data] {
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeNode(b, c)
		}
		b.WriteString("</")
		b.WriteString(n.Data)
		b.WriteString(">")
	default:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeNode(b, c)
		}
	}
}

// voidElements are the elements without an end tag.
var voidElements = map[string]bool{
	"area":   true,
	"base":   true,
	"br":     true,
	"col":    true,
	"embed":  true,
	"hr":     true,
	"img":    true,
	"input":  true,
	"link":   true,
	"meta":   true,
	"source": true,
	"track":  true,
	"wbr":    true,
}
