package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

func GenerateGridCSS(inputHTML string) (string, error) {
	type gridTemplate struct {
		Columns []string
		Rows    []string
		Areas   [][]string
	}
	type gridNode struct {
		Node     *html.Node
		Selector string
		Items    []string
		Base     gridTemplate
		BP       map[string]gridTemplate
	}

	breakpointOrder := []string{"sm", "md", "lg", "xl", "xxl"}
	breakpointMinWidth := map[string]string{
		"sm":  "576px",
		"md":  "768px",
		"lg":  "992px",
		"xl":  "1200px",
		"xxl": "1400px",
	}

	reservedKeywords := map[string]struct{}{
		"span":         {},
		"auto":         {},
		"inherit":      {},
		"initial":      {},
		"unset":        {},
		"revert":       {},
		"revert-layer": {},
	}

	kebabRE := regexp.MustCompile(`^[a-z]+(?:-[a-z0-9]+)*$`)
	onlyDigitsRE := regexp.MustCompile(`^\d+$`)
	frRE := regexp.MustCompile(`^[1-9]\d*fr$`)
	minmaxRE := regexp.MustCompile(`^minmax\(\s*([^,]+)\s*,\s*([^)]+)\s*\)$`)

	getAttr := func(n *html.Node, key string) (string, bool) {
		for _, a := range n.Attr {
			if a.Key == key {
				return a.Val, true
			}
		}
		return "", false
	}

	parseSpaceSeparated := func(raw string) []string {
		return strings.Fields(strings.TrimSpace(raw))
	}

	isAllowedSize := func(v string) bool {
		v = strings.TrimSpace(v)
		if v == "min-content" || v == "max-content" {
			return true
		}
		if frRE.MatchString(v) {
			return true
		}
		m := minmaxRE.FindStringSubmatch(v)
		if len(m) == 3 {
			a := strings.TrimSpace(m[1])
			b := strings.TrimSpace(m[2])
			okPart := func(x string) bool {
				return x == "min-content" || x == "max-content" || frRE.MatchString(x)
			}
			return okPart(a) && okPart(b)
		}
		return false
	}

	validateSizes := func(paramName string, values []string, emptyErr string) error {
		if len(values) == 0 {
			return fmt.Errorf("%s: The parameter **%s** cannot be an empty array.", emptyErr, paramName)
		}
		for _, v := range values {
			if !isAllowedSize(v) {
				return fmt.Errorf("Unsupported size unit: The value **%s** uses an unsupported size unit.", v)
			}
		}
		return nil
	}

	validateAreas := func(paramName string, areas [][]string) error {
		if len(areas) == 0 {
			return fmt.Errorf("Empty grid template area: The parameter **%s** cannot be empty.", paramName)
		}
		for _, row := range areas {
			if len(row) == 0 {
				return fmt.Errorf("Empty grid template area: The parameter **%s** cannot be empty.", paramName)
			}
		}
		return nil
	}

	validateGridAreaName := func(area string) error {
		if _, found := reservedKeywords[area]; found {
			return fmt.Errorf("Reserved CSS keyword: The grid area **%s** uses a reserved CSS keyword and cannot be used.", area)
		}
		if onlyDigitsRE.MatchString(area) {
			return fmt.Errorf("Invalid built-in arguments: The parameter **items** does not match the schema of the built-in component **WEB_GRID**.")
		}
		if !kebabRE.MatchString(area) {
			return fmt.Errorf("Invalid built-in arguments: The parameter **items** does not match the schema of the built-in component **WEB_GRID**.")
		}
		return nil
	}

	validateRectangles := func(paramName string, areas [][]string) error {
		type point struct{ r, c int }
		positions := map[string][]point{}

		for r := range areas {
			for c := range areas[r] {
				name := areas[r][c]
				if name == "." {
					continue
				}
				positions[name] = append(positions[name], point{r: r, c: c})
			}
		}

		for area, pts := range positions {
			if len(pts) == 0 {
				continue
			}
			minR, maxR := pts[0].r, pts[0].r
			minC, maxC := pts[0].c, pts[0].c
			set := map[[2]int]struct{}{}
			for _, p := range pts {
				set[[2]int{p.r, p.c}] = struct{}{}
				if p.r < minR {
					minR = p.r
				}
				if p.r > maxR {
					maxR = p.r
				}
				if p.c < minC {
					minC = p.c
				}
				if p.c > maxC {
					maxC = p.c
				}
			}

			rectCount := (maxR - minR + 1) * (maxC - minC + 1)
			if rectCount != len(pts) {
				visited := map[[2]int]struct{}{}
				queue := [][2]int{{pts[0].r, pts[0].c}}
				visited[[2]int{pts[0].r, pts[0].c}] = struct{}{}

				dirs := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
				for len(queue) > 0 {
					cur := queue[0]
					queue = queue[1:]
					for _, d := range dirs {
						nr, nc := cur[0]+d[0], cur[1]+d[1]
						key := [2]int{nr, nc}
						if _, ok := set[key]; !ok {
							continue
						}
						if _, ok := visited[key]; ok {
							continue
						}
						visited[key] = struct{}{}
						queue = append(queue, key)
					}
				}

				if len(visited) != len(pts) {
					return fmt.Errorf("Multiple grid area shapes: The grid area **%s** in **%s** creates multiple separate shapes.", area, paramName)
				}
				return fmt.Errorf("Invalid grid area shape: The grid area **%s** in **%s** must form a rectangle.", area, paramName)
			}
		}

		return nil
	}

	validateTemplate := func(columnsName, rowsName, areasName string, t gridTemplate, knownAreas map[string]struct{}) error {
		if err := validateSizes(columnsName, t.Columns, "Empty grid template columns"); err != nil {
			return err
		}
		if err := validateSizes(rowsName, t.Rows, "Empty grid template rows"); err != nil {
			return err
		}
		if err := validateAreas(areasName, t.Areas); err != nil {
			return err
		}
		if len(t.Areas) != len(t.Rows) {
			return fmt.Errorf("Unmatched rows: The number of rows in **%s** does not match **%s**.", areasName, rowsName)
		}
		for _, row := range t.Areas {
			if len(row) != len(t.Columns) {
				return fmt.Errorf("Unmatched columns: The number of columns in **%s** does not match **%s**.", areasName, columnsName)
			}
			for _, area := range row {
				if area == "." {
					continue
				}
				if _, ok := knownAreas[area]; !ok {
					return fmt.Errorf("Unknown grid area: The grid area **%s** is used in **%s** but is not defined in **items**.", area, areasName)
				}
			}
		}
		return validateRectangles(areasName, t.Areas)
	}

	templateUsesArea := func(t gridTemplate, area string) bool {
		for _, row := range t.Areas {
			for _, cell := range row {
				if cell == area {
					return true
				}
			}
		}
		return false
	}

	areasToCSSValue := func(areas [][]string) string {
		lines := make([]string, 0, len(areas))
		for _, row := range areas {
			lines = append(lines, `"`+strings.Join(row, " ")+`"`)
		}
		return strings.Join(lines, " ")
	}

	parseAreasAttr := func(raw string) ([][]string, error) {
		var out [][]string
		if err := json.Unmarshal([]byte(raw), &out); err != nil {
			return nil, fmt.Errorf("Invalid built-in arguments: The parameter **grid-template-areas** does not match the schema of the built-in component **WEB_GRID**.")
		}
		return out, nil
	}

	parseTemplateFromNode := func(n *html.Node, prefix string) (gridTemplate, bool, error) {
		var colsKey, rowsKey, areasKey string
		if prefix == "" {
			colsKey = "data-grid-template-columns"
			rowsKey = "data-grid-template-rows"
			areasKey = "data-grid-template-areas"
		} else {
			colsKey = "data-" + prefix + "-grid-template-columns"
			rowsKey = "data-" + prefix + "-grid-template-rows"
			areasKey = "data-" + prefix + "-grid-template-areas"
		}

		colsVal, hasCols := getAttr(n, colsKey)
		rowsVal, hasRows := getAttr(n, rowsKey)
		areasVal, hasAreas := getAttr(n, areasKey)

		if prefix != "" {
			if hasCols || hasRows || hasAreas {
				if !(hasCols && hasRows && hasAreas) {
					return gridTemplate{}, false, fmt.Errorf("Missing breakpoint grid template parameters: The breakpoint **%s** must define **grid-template-columns**, **grid-template-rows**, and **grid-template-areas** together.", prefix)
				}
			} else {
				return gridTemplate{}, false, nil
			}
		} else {
			if !(hasCols && hasRows && hasAreas) {
				return gridTemplate{}, false, fmt.Errorf("Invalid built-in arguments: The parameter **grid-template-columns** does not match the schema of the built-in component **WEB_GRID**.")
			}
		}

		areas, err := parseAreasAttr(areasVal)
		if err != nil {
			return gridTemplate{}, false, err
		}

		return gridTemplate{
			Columns: parseSpaceSeparated(colsVal),
			Rows:    parseSpaceSeparated(rowsVal),
			Areas:   areas,
		}, true, nil
	}

	root, err := html.Parse(strings.NewReader("<!doctype html><html><body>" + inputHTML + "</body></html>"))
	if err != nil {
		return "", err
	}

	var buildSelector func(*html.Node) string
	buildSelector = func(n *html.Node) string {
		var parts []string
		cur := n
		for cur != nil && !(cur.Type == html.ElementNode && cur.Data == "body") {
			if cur.Type == html.ElementNode {
				index := 0
				for sib := cur.PrevSibling; sib != nil; sib = sib.PrevSibling {
					if sib.Type == html.ElementNode && sib.Data == cur.Data {
						index++
					}
				}
				parts = append([]string{fmt.Sprintf("%s:nth-of-type(%d)", cur.Data, index+1)}, parts...)
			}
			cur = cur.Parent
		}
		return strings.Join(parts, " > ")
	}

	var grids []gridNode
	var walk func(*html.Node) error
	walk = func(n *html.Node) error {
		if n.Type == html.ElementNode && n.Data == "compono-web-grid" {
			var items []string
			seen := map[string]struct{}{}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "compono-web-grid-item" {
					area, ok := getAttr(c, "data-grid-area")
					if !ok {
						return fmt.Errorf("Invalid built-in arguments: The parameter **items** does not match the schema of the built-in component **WEB_GRID**.")
					}
					if err := validateGridAreaName(area); err != nil {
						return err
					}
					if _, exists := seen[area]; exists {
						return fmt.Errorf("Grid areas must be unique: The grid area **%s** is used more than once in **items**.", area)
					}
					seen[area] = struct{}{}
					items = append(items, area)
				}
			}
			if len(items) == 0 {
				return fmt.Errorf("Empty items: The parameter **items** cannot be an empty array.")
			}

			knownAreas := map[string]struct{}{}
			for _, a := range items {
				knownAreas[a] = struct{}{}
			}

			base, _, err := parseTemplateFromNode(n, "")
			if err != nil {
				return err
			}
			if err := validateTemplate("grid-template-columns", "grid-template-rows", "grid-template-areas", base, knownAreas); err != nil {
				return err
			}

			bpMap := map[string]gridTemplate{}
			for _, bp := range breakpointOrder {
				t, ok, err := parseTemplateFromNode(n, bp)
				if err != nil {
					return err
				}
				if ok {
					if err := validateTemplate(bp+"-grid-template-columns", bp+"-grid-template-rows", bp+"-grid-template-areas", t, knownAreas); err != nil {
						return err
					}
					bpMap[bp] = t
				}
			}

			usedAnywhere := map[string]bool{}
			for _, area := range items {
				if templateUsesArea(base, area) {
					usedAnywhere[area] = true
				}
			}
			for _, bp := range breakpointOrder {
				if t, ok := bpMap[bp]; ok {
					for _, area := range items {
						if templateUsesArea(t, area) {
							usedAnywhere[area] = true
						}
					}
				}
			}
			for _, area := range items {
				if !usedAnywhere[area] {
					return fmt.Errorf("Unused grid area: The grid area **%s** is defined in **items** but is not used in any grid template areas.", area)
				}
			}

			grids = append(grids, gridNode{
				Node:     n,
				Selector: buildSelector(n),
				Items:    items,
				Base:     base,
				BP:       bpMap,
			})
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := walk(c); err != nil {
				return err
			}
		}
		return nil
	}

	if err := walk(root); err != nil {
		return "", err
	}

	var css bytes.Buffer
	css.WriteString("compono-web-grid{display:grid;}\n")

	for _, g := range grids {
		css.WriteString(fmt.Sprintf("%s{display:grid;grid-template-columns:%s;grid-template-rows:%s;grid-template-areas:%s;}\n",
			g.Selector,
			strings.Join(g.Base.Columns, " "),
			strings.Join(g.Base.Rows, " "),
			areasToCSSValue(g.Base.Areas),
		))

		usedBase := map[string]bool{}
		for _, row := range g.Base.Areas {
			for _, cell := range row {
				if cell != "." {
					usedBase[cell] = true
				}
			}
		}

		for _, area := range g.Items {
			if usedBase[area] {
				css.WriteString(fmt.Sprintf("%s > compono-web-grid-item[data-grid-area=\"%s\"]{grid-area:%s;display:block;}\n", g.Selector, area, area))
			} else {
				css.WriteString(fmt.Sprintf("%s > compono-web-grid-item[data-grid-area=\"%s\"]{display:none;}\n", g.Selector, area))
			}
		}

		current := g.Base
		for _, bp := range breakpointOrder {
			if t, ok := g.BP[bp]; ok {
				current = t
			}

			used := map[string]bool{}
			for _, row := range current.Areas {
				for _, cell := range row {
					if cell != "." {
						used[cell] = true
					}
				}
			}

			css.WriteString(fmt.Sprintf("@media (min-width:%s){%s{display:grid;grid-template-columns:%s;grid-template-rows:%s;grid-template-areas:%s;}",
				breakpointMinWidth[bp],
				g.Selector,
				strings.Join(current.Columns, " "),
				strings.Join(current.Rows, " "),
				areasToCSSValue(current.Areas),
			))

			areasSorted := append([]string(nil), g.Items...)
			sort.Strings(areasSorted)
			for _, area := range areasSorted {
				if used[area] {
					css.WriteString(fmt.Sprintf("%s > compono-web-grid-item[data-grid-area=\"%s\"]{grid-area:%s;display:block;}", g.Selector, area, area))
				} else {
					css.WriteString(fmt.Sprintf("%s > compono-web-grid-item[data-grid-area=\"%s\"]{display:none;}", g.Selector, area))
				}
			}
			css.WriteString("}\n")
		}
	}

	return css.String(), nil
}
