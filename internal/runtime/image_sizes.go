package runtime

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

func OptimizeHTML(inputHTML string) (string, error) {
	root, err := parseHTMLFragment(inputHTML)
	if err != nil {
		return "", err
	}

	applyImageSizes(root)

	return renderHTMLFragment(root), nil
}

func parseHTMLFragment(inputHTML string) (*html.Node, error) {
	root, err := html.Parse(strings.NewReader("<!doctype html><html><body>" + inputHTML + "</body></html>"))
	if err != nil {
		return nil, err
	}

	body := findHTMLElement(root, "body")
	if body == nil {
		return &html.Node{Type: html.ElementNode, Data: "body"}, nil
	}

	return body, nil
}

func renderHTMLFragment(root *html.Node) string {
	var buf bytes.Buffer
	for node := root.FirstChild; node != nil; node = node.NextSibling {
		_ = html.Render(&buf, node)
	}
	return buf.String()
}

func applyImageSizes(root *html.Node) {
	processedPictures := map[*html.Node]struct{}{}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			picture := closestAncestor(n, "picture")
			if picture != nil {
				if _, ok := processedPictures[picture]; !ok {
					processedPictures[picture] = struct{}{}
					applySizesToPicture(picture)
				}
			} else if _, ok := getHTMLAttr(n, "srcset"); ok {
				setHTMLAttr(n, "sizes", "100vw")
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
}

func applySizesToPicture(picture *html.Node) {
	for child := picture.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode {
			continue
		}
		if child.Data == "source" || child.Data == "img" {
			if _, ok := getHTMLAttr(child, "srcset"); ok {
				setHTMLAttr(child, "sizes", "100vw")
			}
		}
	}
}

func closestAncestor(n *html.Node, tag string) *html.Node {
	for parent := n.Parent; parent != nil; parent = parent.Parent {
		if parent.Type == html.ElementNode && parent.Data == tag {
			return parent
		}
	}
	return nil
}

func findHTMLElement(root *html.Node, tag string) *html.Node {
	if root.Type == html.ElementNode && root.Data == tag {
		return root
	}

	for child := root.FirstChild; child != nil; child = child.NextSibling {
		if found := findHTMLElement(child, tag); found != nil {
			return found
		}
	}

	return nil
}

func getHTMLAttr(n *html.Node, key string) (string, bool) {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val, true
		}
	}
	return "", false
}

func setHTMLAttr(n *html.Node, key, value string) {
	for i := range n.Attr {
		if n.Attr[i].Key == key {
			n.Attr[i].Val = value
			return
		}
	}

	n.Attr = append(n.Attr, html.Attribute{Key: key, Val: value})
}
