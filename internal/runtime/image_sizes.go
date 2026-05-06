package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type imageGridTemplate struct {
	Columns []string
	Areas   [][]string
}

type imageGrid struct {
	Base        imageGridTemplate
	Breakpoints map[string]imageGridTemplate
}

var imageBreakpointOrder = []string{"sm", "md", "lg", "xl", "xxl"}

var imageBreakpointMinWidth = map[string]string{
	"sm":  "576px",
	"md":  "768px",
	"lg":  "992px",
	"xl":  "1200px",
	"xxl": "1400px",
}

var imageFRRE = regexp.MustCompile(`^([1-9]\d*)fr$`)

func OptimizeHTML(inputHTML string) (string, error) {
	root, err := parseHTMLFragment(inputHTML)
	if err != nil {
		return "", err
	}

	grids := collectImageGrids(root)
	applyImageSizes(root, grids)

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

func collectImageGrids(root *html.Node) map[*html.Node]imageGrid {
	grids := map[*html.Node]imageGrid{}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "compono-web-grid" {
			if grid, ok := parseImageGrid(n); ok {
				grids[n] = grid
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)

	return grids
}

func parseImageGrid(n *html.Node) (imageGrid, bool) {
	base, ok := parseImageGridTemplate(n, "")
	if !ok {
		return imageGrid{}, false
	}

	grid := imageGrid{
		Base:        base,
		Breakpoints: map[string]imageGridTemplate{},
	}
	for _, bp := range imageBreakpointOrder {
		if template, ok := parseImageGridTemplate(n, bp); ok {
			grid.Breakpoints[bp] = template
		}
	}

	return grid, true
}

func parseImageGridTemplate(n *html.Node, prefix string) (imageGridTemplate, bool) {
	colsKey := "data-grid-template-columns"
	areasKey := "data-grid-template-areas"
	if prefix != "" {
		colsKey = "data-" + prefix + "-grid-template-columns"
		areasKey = "data-" + prefix + "-grid-template-areas"
	}

	colsVal, hasCols := getHTMLAttr(n, colsKey)
	areasVal, hasAreas := getHTMLAttr(n, areasKey)
	if !hasCols || !hasAreas {
		return imageGridTemplate{}, false
	}

	var areas [][]string
	if err := json.Unmarshal([]byte(areasVal), &areas); err != nil {
		return imageGridTemplate{}, false
	}

	return imageGridTemplate{
		Columns: strings.Fields(strings.TrimSpace(colsVal)),
		Areas:   areas,
	}, true
}

func applyImageSizes(root *html.Node, grids map[*html.Node]imageGrid) {
	processedPictures := map[*html.Node]struct{}{}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			picture := closestAncestor(n, "picture")
			if picture != nil {
				if _, ok := processedPictures[picture]; !ok {
					processedPictures[picture] = struct{}{}
					applySizesToPicture(picture, imageSizesForNode(picture, grids))
				}
			} else if _, ok := getHTMLAttr(n, "srcset"); ok {
				setHTMLAttr(n, "sizes", imageSizesForNode(n, grids))
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
}

func applySizesToPicture(picture *html.Node, sizes string) {
	for child := picture.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode {
			continue
		}
		if child.Data == "source" {
			if _, ok := getHTMLAttr(child, "srcset"); ok {
				setHTMLAttr(child, "sizes", sizes)
			}
		}
		if child.Data == "img" {
			if _, ok := getHTMLAttr(child, "srcset"); ok {
				setHTMLAttr(child, "sizes", sizes)
			}
		}
	}
}

func imageSizesForNode(n *html.Node, grids map[*html.Node]imageGrid) string {
	item := closestAncestor(n, "compono-web-grid-item")
	if item == nil || item.Parent == nil || item.Parent.Data != "compono-web-grid" {
		return "100vw"
	}

	area, ok := getHTMLAttr(item, "data-grid-area")
	if !ok {
		return "100vw"
	}

	grid, ok := grids[item.Parent]
	if !ok {
		return "100vw"
	}

	return imageSizesForGridArea(grid, area)
}

func imageSizesForGridArea(grid imageGrid, area string) string {
	base := imageTemplateAreaSize(grid.Base, area)
	if base == "" {
		base = "100vw"
	}

	currentTemplate := grid.Base
	currentSize := base
	changes := make([]string, 0, len(imageBreakpointOrder))

	for _, bp := range imageBreakpointOrder {
		if template, ok := grid.Breakpoints[bp]; ok {
			currentTemplate = template
		}

		size := imageTemplateAreaSize(currentTemplate, area)
		if size == "" {
			size = "100vw"
		}
		if size != currentSize {
			changes = append(changes, fmt.Sprintf("(min-width: %s) %s", imageBreakpointMinWidth[bp], size))
			currentSize = size
		}
	}

	parts := make([]string, 0, len(changes)+1)
	for i := len(changes) - 1; i >= 0; i-- {
		parts = append(parts, changes[i])
	}
	parts = append(parts, base)
	return strings.Join(parts, ", ")
}

func imageTemplateAreaSize(template imageGridTemplate, area string) string {
	minCol, maxCol, ok := imageAreaColumns(template.Areas, area)
	if !ok || minCol < 0 || maxCol >= len(template.Columns) {
		return ""
	}

	totalFR := 0
	spanFR := 0
	for i, column := range template.Columns {
		fr, ok := parseImageFR(column)
		if !ok {
			return "100vw"
		}
		totalFR += fr
		if i >= minCol && i <= maxCol {
			spanFR += fr
		}
	}
	if totalFR == 0 || spanFR == 0 {
		return ""
	}

	return formatViewportWidth(float64(spanFR) / float64(totalFR) * 100)
}

func imageAreaColumns(areas [][]string, area string) (int, int, bool) {
	minCol := math.MaxInt
	maxCol := -1

	for _, row := range areas {
		for col, cell := range row {
			if cell != area {
				continue
			}
			if col < minCol {
				minCol = col
			}
			if col > maxCol {
				maxCol = col
			}
		}
	}

	if maxCol == -1 {
		return 0, 0, false
	}
	return minCol, maxCol, true
}

func parseImageFR(value string) (int, bool) {
	matches := imageFRRE.FindStringSubmatch(strings.TrimSpace(value))
	if len(matches) != 2 {
		return 0, false
	}

	fr, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, false
	}
	return fr, true
}

func formatViewportWidth(value float64) string {
	rounded := math.Round(value)
	if math.Abs(value-rounded) < 0.005 {
		return strconv.FormatInt(int64(rounded), 10) + "vw"
	}

	formatted := strconv.FormatFloat(value, 'f', 2, 64)
	formatted = strings.TrimRight(strings.TrimRight(formatted, "0"), ".")
	return formatted + "vw"
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
