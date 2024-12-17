package latex2mathml

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/neruyzo/etree"
	"github.com/wanglihui/pptx-go/latex2mathml/slices"
)

var COLUMNS_ALINGNMENT_MAP = map[string]string{"r": "right", "l": "left", "c": "center"}
var SPACE_LIST = []string{`\ `, "~", NOBREAKSPACE, SPACE}
var SUB_LIST = []string{DETERMINANT, GCD, INTOP, INJLIM, LIMINF, LIMSUP, PR, PROJLIM}

var OPERATORS = []string{
	"+",
	"-",
	"*",
	"/",
	"(",
	")",
	"=",
	",",
	"?",
	"[",
	"]",
	"|",
	`\|`,
	"!",
	`\{`,
	`\}`,
	`>`,
	`<`,
	`.`,
	`\bigotimes`,
	`\centerdot`,
	`\dots`,
	`\dotsc`,
	`\dotso`,
	`\gt`,
	`\ldotp`,
	`\lt`,
	`\lvert`,
	`\lVert`,
	`\lvertneqq`,
	`\ngeqq`,
	`\omicron`,
	`\rvert`,
	`\rVert`,
	`\S`,
	`\smallfrown`,
	`\smallint`,
	`\smallsmile`,
	`\surd`,
	`\varsubsetneqq`,
	`\varsupsetneqq`,
}

type Mode int64

const (
	TextMode Mode = 0
	MathMode Mode = 1
)

type Moded struct {
	Value string
	Mode  Mode
}

var MATH_MODE_PATTERN = regexp.MustCompile(`\\\$|\$|\\?[^\\$]+`)
var NUMBER_PATTERN = regexp.MustCompile(`\d+(.\d+)?`)

func Convert(latex string, xmlns string, display string, indent int) string {
	InitializeCommands()
	ParseSymbol()
	doc := etree.NewDocument()
	math := doc.CreateElement("math")
	math.CreateAttr("xmlns", xmlns)
	math.CreateAttr("display", display)
	row := math.CreateElement("mrow")
	nodes, _ := Walk(latex)
	convertGroup(nodes, row, map[string]string{})
	if indent != 0 {
		doc.Indent(indent)
	}
	doc.WriteSettings.NoEscape = true
	mathml, _ := doc.WriteToString()
	return mathml
}

func convertMatrix(nodes []Node, parent *etree.Element, command string, alignment string) {
	var row *etree.Element
	var cell *etree.Element

	var colIndex = 0
	var maxColSize = 0
	var colAlignment *string

	var rowIndex = 0
	var rowLines = []string{}

	var indexes = []bool{}

	for _, node := range nodes {

		if row == nil {
			row = parent.CreateElement("mtr")
		}

		if cell == nil {
			colAlignment, colIndex = getColumnAlignment(&alignment, colAlignment, colIndex)
			cell = makeMatrixCell(row, colAlignment)
		}

		if node.Token == BRACES {
			convertGroup([]Node{node}, cell, map[string]string{})
		} else if node.Token == "&" {
			setCellAlignment(cell, indexes)
			indexes = []bool{}
			colAlignment, colIndex = getColumnAlignment(&alignment, colAlignment, colIndex)
			cell = makeMatrixCell(row, colAlignment)
			if (command == SPLIT || command == ALIGN) && colIndex%2 == 0 {
				cell.CreateElement("mi")
			}
		} else if node.Token == DOUBLEBACKSLASH || node.Token == CARRIAGE_RETURN {
			setCellAlignment(cell, indexes)
			indexes = []bool{}
			rowIndex = rowIndex + 1
			if colIndex > maxColSize {
				maxColSize = colIndex
			}
			colIndex = 0
			colAlignment, colIndex = getColumnAlignment(&alignment, colAlignment, colIndex)
			row = parent.CreateElement("mtr")
			cell = makeMatrixCell(row, colAlignment)
		} else if node.Token == HLINE {
			rowLines = append(rowLines, "solid")
		} else if node.Token == HDASHLINE {
			rowLines = append(rowLines, "dashed")
		} else if node.Token == HFIL {
			indexes = append(indexes, true)
		} else {
			if rowIndex > len(rowLines) {
				rowLines = append(rowLines, "none")
			}
			indexes = append(indexes, false)
			convertGroup([]Node{node}, cell, map[string]string{})
		}
	}

	if colIndex > maxColSize {
		maxColSize = colIndex
	}

	if slices.Contains(rowLines, "solid") {
		parent.CreateAttr("rowlines", strings.Join(rowLines, " "))
	}

	if row != nil && cell != nil && len(cell.ChildElements()) == 0 && cell.Text() == "" {
		children := parent.ChildElements()
		parent.RemoveChildAt(len(children) - 1)
	}

	if maxColSize > 0 && command == ALIGN {
		multiplier := maxColSize / 2
		spacing := "0em 2em"
		for i := 0; i < multiplier-1; i++ {
			spacing = spacing + " 0em 2em"
		}
		parent.CreateAttr("columnspacing", spacing)
	}
}

func setCellAlignment(cell *etree.Element, indexes []bool) {
	if slices.Contains(indexes, true) {
		if indexes[0] && !indexes[len(indexes)-1] {
			cell.CreateAttr("columnalign", "right")
		} else if !indexes[0] && indexes[len(indexes)-1] {
			cell.CreateAttr("columnalign", "left")
		}
	}
}

func getColumnAlignment(alignment *string, columnAlignment *string, columnIndex int) (*string, int) {
	if alignment != nil && *alignment != "" {
		var align = *alignment
		var alignmentIndex = align[columnIndex%len(align)]
		columnAlign, exist := COLUMNS_ALINGNMENT_MAP[string(alignmentIndex)]

		if exist {
			columnAlignment = &columnAlign
		}

		columnIndex = columnIndex + 1
	}

	return columnAlignment, columnIndex
}

func makeMatrixCell(row *etree.Element, columnAlignment *string) *etree.Element {
	var element = etree.NewElement("mtd")

	if columnAlignment != nil {
		element.CreateAttr("columnalign", *columnAlignment)
	}
	row.AddChild(element)

	return element
}

func convertGroup(nodes []Node, parent *etree.Element, font map[string]string) {
	for _, node := range nodes {
		token := node.Token

		if _, exist := MSTYLE_SIZES[token]; exist {
			node := Node{Token: token, Children: nodes}
			convertCommand(node, parent, font)
		} else if _, exist := STYLES[token]; exist {
			node := Node{Token: token, Children: nodes}
			convertCommand(node, parent, font)
		} else if _, exist := CONVERSION_MAP[token]; exist || token == MOD || token == PMOD {
			convertCommand(node, parent, font)
		} else if _, exist := LOCAL_FONTS[token]; exist && node.Children != nil {
			convertGroup(node.Children, parent, LOCAL_FONTS[token])
		} else if strings.HasPrefix(token, MATH) && node.Children != nil {
			convertGroup(node.Children, parent, font)
		} else if _, exist := GLOBAL_FONTS[token]; exist {
			font, _ = GLOBAL_FONTS[token]
		} else if node.Children == nil {
			convertSymbol(node, parent, font)
		} else {
			row := parent.CreateElement("mrow")
			addAttributes(row, node.Attributes)
			convertGroup(node.Children, row, font)
		}
	}

}

func getAlignmentAndColumnLine(alignment *string) (*string, *string) {
	if alignment == nil {
		return nil, nil
	}

	if !strings.Contains(*alignment, "|") {
		return alignment, nil
	}

	var ajusted = ""
	var columnLines = []string{}

	for _, char := range *alignment {
		if char == '|' {
			columnLines = append(columnLines, "solid")
		} else {
			ajusted = ajusted + string(char)
		}

		if len(ajusted)-len(columnLines) == 2 {
			columnLines = append(columnLines, "none")
		}
	}

	var column = strings.Join(columnLines, " ")

	return &ajusted, &column
}

func SeparateByMode(text string) []Moded {
	var value = ""
	var isMathMode = false

	var moded = []Moded{}

	for _, match := range MATH_MODE_PATTERN.FindAllString(text, 0) {
		if match == "$" {
			if isMathMode {
				moded = append(moded, Moded{Value: value, Mode: MathMode})
			} else {
				moded = append(moded, Moded{Value: value, Mode: TextMode})
			}
			value = ""
			isMathMode = !isMathMode
		} else {
			value = value + match
		}
	}

	if len(value) > 0 {
		if isMathMode {
			moded = append(moded, Moded{Value: value, Mode: MathMode})
		} else {
			moded = append(moded, Moded{Value: value, Mode: TextMode})
		}
	}

	return moded
}

func convertCommand(node Node, parent *etree.Element, font map[string]string) {
	command := node.Token
	modifier := node.Modifier

	if command == SUBSTACK || command == SMALLMATRIX {
		parent = parent.CreateElement("mstyle")
		parent.CreateAttr("scriptlevel", "1")
	} else if command == CASES {
		parent = parent.CreateElement("mrow")
		lbrace := parent.CreateElement("mo")
		lbrace.CreateAttr("stretchy", "true")
		lbrace.CreateAttr("fence", "true")
		lbrace.CreateAttr("form", "prefix")
		code, _ := ConvertSymbol(LBRACE)
		lbrace.SetText("&#x" + code + ";")
	} else if command == DBINOM || command == DFRAC {
		parent = parent.CreateElement("mstyle")
		parent.CreateAttr("displaystyle", "true")
		parent.CreateAttr("scriptlevel", "0")
	} else if command == HPHANTOM {
		parent = parent.CreateElement("mpadded")
		parent.CreateAttr("height", "0")
		parent.CreateAttr("depth", "0")
	} else if command == VPHANTOM {
		parent = parent.CreateElement("mpadded")
		parent.CreateAttr("width", "0")
	} else if command == TBINOM || command == HBOX || command == MBOX || command == TFRAC {
		parent = parent.CreateElement("mstyle")
		parent.CreateAttr("displaystyle", "false")
		parent.CreateAttr("scriptlevel", "0")
	} else if command == MOD || command == PMOD {
		parent = parent.CreateElement("mspace")
		parent.CreateAttr("width", "1em")
	}

	style, _ := CONVERSION_MAP[command]

	if len(node.Attributes) > 0 && node.Token != SKEW {
		for key, value := range node.Attributes {
			style.Modifiers[key] = value
		}
	}

	if command == LEFT {
		parent = parent.CreateElement("mrow")
	}

	appendPrefixElement(node, parent)

	alignment, columnLines := getAlignmentAndColumnLine(&node.Alignment)

	if columnLines != nil {
		style.Modifiers["columnlines"] = *columnLines
	}

	var tag = style.Tag

	if command == SUBSUP && len(node.Children) > 0 && node.Children[0].Token == GCD {
		tag = "munderover"
	} else if command == SUPERSCRIPT && (modifier == LIMITS || modifier == OVERBRACE) {
		tag = "mover"
	} else if command == SUBSCRIPT && (modifier == LIMITS || modifier == UNDERBRACE) {
		tag = "munder"
	} else if command == SUBSUP && (modifier == LIMITS || modifier == OVERBRACE || modifier == UNDERBRACE) {
		tag = "munderover"
	} else if (command == XLEFTARROW || command == XRIGHTARROW) && len(node.Children) == 2 {
		tag = "munderover"
	}

	element := parent.CreateElement(tag)
	addAttributes(element, style.Modifiers)

	if slices.Contains(LIMIT, command) {
		element.SetText(command[1:])
	} else if command == MOD || command == PMOD {
		element.SetText("mod")
		space := parent.CreateElement("mspace")
		space.CreateAttr("width", "0.333em")
	} else if command == BMOD {
		element.SetText("mod")
	} else if command == XLEFTARROW || command == XRIGHTARROW {
		style := element.CreateElement("mstyle")
		style.CreateAttr("scriptlevel", "0")
		arrow := style.CreateElement("mo")
		if command == XLEFTARROW {
			arrow.SetText("&#x2190;")
		} else {
			arrow.SetText("&#x2192;")
		}
	} else if node.Text != "" {
		if command == MIDDLE {
			code, _ := ConvertSymbol(node.Text)
			element.SetText("&#x" + code + ";")
		} else if command == HBOX {
			mtext := element
			for _, mode := range SeparateByMode(node.Text) {
				if mode.Mode == TextMode {
					if mtext == nil {
						mtext = parent.CreateElement(tag)
						addAttributes(mtext, style.Modifiers)
						setFont(mtext, "mtext", font)
						mtext = nil
					} else {
						row := parent.CreateElement("mrow")
						nodes, _ := Walk(mode.Value)
						convertGroup(nodes, row, map[string]string{})
					}
				}
			}
		} else {
			if command == FBOX {
				element = element.CreateElement("mtext")
			}
			element.SetText(strings.ReplaceAll(node.Text, " ", "&#x000A0;"))
			setFont(element, "mtext", font)
		}

	} else if node.Delimiter != "" && command != FRAC && command != GENFRAC {

		if node.Delimiter != "." {
			code, err := ConvertSymbol(node.Delimiter)
			if err == nil {
				element.SetText("&#x" + code + ";")
			}
		}
	}

	if node.Children != nil {
		localParent := element

		if command == LEFT || command == MOD || command == PMOD {
			localParent = parent
		}

		align := *alignment

		if slices.Contains(MATRICES, command) {

			if command == CASES {
				align = "l"
			} else if command == SPLIT || command == ALIGN {
				align = "rl"
			}
			convertMatrix(node.Children, localParent, command, align)

		} else if command == CFRAC {

			for _, child := range node.Children {
				p := localParent.CreateElement("mstyle")
				p.CreateAttr("displaystyle", "false")
				p.CreateAttr("scriptlevel", "0")
				convertGroup([]Node{child}, p, font)
			}

		} else if command == SIDESET {

			convertGroup(node.Children[0:1], localParent, font)
			fill := localParent.CreateElement("mstyle")
			fill.CreateAttr("scriptlevel", "0")
			space := fill.CreateElement("mspace")
			space.CreateAttr("width", "-0.167em")
			convertGroup(node.Children[1:2], localParent, font)

		} else if command == SKEW {

			child := node.Children[0]
			newNode := Node{
				Token: child.Token,
				Children: []Node{
					{
						Token: BRACES,
						Children: append(
							child.Children,
							Node{Token: MKERN, Attributes: node.Attributes},
						),
					},
				},
			}
			convertGroup([]Node{newNode}, localParent, font)

		} else if command == XLEFTARROW || command == XRIGHTARROW {
			for _, child := range node.Children {
				padded := localParent.CreateElement("mpadded")
				// padded.CreateAttr("width", "0.833em")
				// padded.CreateAttr("lspace", "0.556em")
				// padded.CreateAttr("voffset", "-0.2em")
				// padded.CreateAttr("height", "-0.2em")
				convertGroup([]Node{child}, padded, font)
				space := padded.CreateElement("mspace")
				space.CreateAttr("depth", "0.25em")
			}

		} else {

			convertGroup(node.Children, localParent, font)
		}
	}

	addDiacritic(command, element)
	appendPostfixElement(node, parent)
}

func addDiacritic(command string, parent *etree.Element) {
	style, exist := DIACRITICS[command]
	if exist {
		element := etree.NewElement("mo")
		element.SetText(style.Tag)
		for key, value := range style.Modifiers {
			element.CreateAttr(key, value)
		}
		parent.AddChild(element)
	}
}

func addAttributes(element *etree.Element, attributes map[string]string) {
	for key, value := range attributes {
		element.CreateAttr(key, value)
	}
}

func convertAndAppendCommand(command string, parent *etree.Element, attributes map[string]string) {
	code, err := ConvertSymbol(command)
	element := parent.CreateElement("mo")
	addAttributes(element, attributes)
	if err == nil {
		element.SetText("&#x" + code + ";")
	} else {
		element.SetText(command)
	}
}

func appendPrefixElement(node Node, parent *etree.Element) {
	var size = "2.047em"

	if parent.SelectAttrValue("displaystyle", "none") == "false" || node.Token == TBINOM {
		size = "1.2em"
	}

	if node.Token == `\pmatrix` || node.Token == PMOD {
		convertAndAppendCommand(`\lparen`, parent, map[string]string{})
	} else if node.Token == BINOM || node.Token == DBINOM || node.Token == TBINOM {
		convertAndAppendCommand(`\lparen`, parent, map[string]string{"minsize": size, "maxsize": size})
	} else if node.Token == `\bmatrix` {
		convertAndAppendCommand(`\lbrack`, parent, map[string]string{})
	} else if node.Token == `\Bmatrix` {
		convertAndAppendCommand(`\lbrace`, parent, map[string]string{})
	} else if node.Token == `\vmatrix` {
		convertAndAppendCommand(`\vert`, parent, map[string]string{})
	} else if node.Token == `\Vmatrix` {
		convertAndAppendCommand(`\Vert`, parent, map[string]string{})
	} else if (node.Token == FRAC || node.Token == GENFRAC) && node.Delimiter != "" && node.Delimiter[0] != '.' {
		convertAndAppendCommand(string(node.Delimiter[0]), parent, map[string]string{"minsize": size, "maxsize": size})
	}
}

func appendPostfixElement(node Node, parent *etree.Element) {
	var size = "2.047em"

	if parent.SelectAttrValue("displaystyle", "none") == "false" || node.Token == TBINOM {
		size = "1.2em"
	}

	if node.Token == `\pmatrix` || node.Token == PMOD {
		convertAndAppendCommand(`\lparen`, parent, map[string]string{})
	} else if node.Token == BINOM || node.Token == DBINOM || node.Token == TBINOM {
		convertAndAppendCommand(`\lparen`, parent, map[string]string{"minsize": size, "maxsize": size})
	} else if node.Token == `\bmatrix` {
		convertAndAppendCommand(`\rbrack`, parent, map[string]string{})
	} else if node.Token == `\Bmatrix` {
		convertAndAppendCommand(`\rbrace`, parent, map[string]string{})
	} else if node.Token == `\vmatrix` {
		convertAndAppendCommand(`\vert`, parent, map[string]string{})
	} else if node.Token == `\Vmatrix` {
		convertAndAppendCommand(`\Vert`, parent, map[string]string{})
	} else if (node.Token == FRAC || node.Token == GENFRAC) && node.Delimiter != "" && node.Delimiter[0] != '.' {
		convertAndAppendCommand(string(node.Delimiter[1]), parent, map[string]string{"minsize": size, "maxsize": size})
	} else if width, ok := node.Attributes["width"]; node.Token == SKEW && ok {
		element := etree.NewElement("mspace")
		element.CreateAttr("width", "-"+width)
		parent.AddChild(element)
	}
}

func convertSymbol(node Node, parent *etree.Element, font map[string]string) {
	token := node.Token
	attributes := node.Attributes
	code, errCode := ConvertSymbol(token)

	if NUMBER_PATTERN.MatchString(token) {

		element := parent.CreateElement("mn")
		addAttributes(element, attributes)
		element.SetText(token)
		setFont(element, element.Tag, font)

	} else if slices.Contains(OPERATORS, token) {

		element := parent.CreateElement("mo")
		addAttributes(element, attributes)
		element.SetText("&#x" + code + ";")

		if token == `\|` {
			element.CreateAttr("fence", "false")
		} else if token == `\smallint` {
			element.CreateAttr("largeop", "false")
		}

		if slices.Contains([]string{"(", ")", "[", "]", "|", `\|`, `\{`, `\}`, `\surd`}, token) {
			element.CreateAttr("stretchy", "false")
			setFont(element, "fence", font)
		} else {
			setFont(element, element.Tag, font)
		}

	} else if value, err := strconv.ParseInt(code, 16, 64); code == "." || err != nil && (value >= 0x2200 && value < 0x22ff+1 || value >= 0x2190 && value < 0x21ff+1) {

		element := parent.CreateElement("mo")
		addAttributes(element, attributes)
		element.SetText("&#x" + code + ";")
		setFont(element, element.Tag, font)

	} else if slices.Contains(SPACE_LIST, token) {

		element := parent.CreateElement("mtext")
		addAttributes(element, attributes)
		element.SetText("&#x000A0;")
		setFont(element, "mtext", font)

	} else if token == NOT {

		padded := parent.CreateElement("mapped")
		padded.CreateAttr("width", "0")
		element := padded.CreateElement("mtext")
		element.SetText("&#x029F8;")

	} else if slices.Contains(SUB_LIST, token) {

		element := parent.CreateElement("mo")
		element.CreateAttr("movablelimits", "true")
		addAttributes(element, attributes)

		if token == INJLIM {
			element.SetText("inj&#x02006;lim")
		} else if token == INTOP {
			element.SetText("&#x0222B;")
		} else if token == LIMINF {
			element.SetText("lim&#x02006;inf")
		} else if token == LIMSUP {
			element.SetText("lim&#x02006;sup")
		} else if token == PROJLIM {
			element.SetText("proj&#x02006;lim")
		} else {
			element.SetText(token[1:])
		}

		setFont(element, element.Tag, font)
	} else if token == IDOTSINT {

		element := parent.CreateElement("mrow")
		addAttributes(element, attributes)

		for _, s := range []string{"&#x0222B;", "&#x022EF;", "&#x0222B;"} {
			child := element.CreateElement("mo")
			child.SetText(s)
		}

	} else if token == LATEX || token == TEX {

		localParent := parent.CreateElement("mrow")
		addAttributes(localParent, attributes)

		if token == LATEX {
			mi_l := localParent.CreateElement("mi")
			mi_l.SetText("L")
			space := localParent.CreateElement("mspace")
			space.CreateAttr("width", "-0.325em")
			padded := localParent.CreateElement("mpadded")
			padded.CreateAttr("height", "0.21ex")
			padded.CreateAttr("depth", "-0.21ex")
			padded.CreateAttr("voffset", "0.21ex")
			style := padded.CreateElement("mstyle")
			style.CreateAttr("displaystyle", "false")
			style.CreateAttr("scriptlevel", "1")
			row := style.CreateElement("mrow")
			mi_a := row.CreateElement("mi")
			mi_a.SetText("A")
			space = localParent.CreateElement("mspace")
			space.CreateAttr("width", "-0.17em")
			setFont(mi_l, mi_l.Tag, font)
			setFont(mi_a, mi_a.Tag, font)
		}

		mi_t := localParent.CreateElement("mi")
		mi_t.SetText("T")
		space := localParent.CreateElement("mspace")
		space.CreateAttr("width", "-0.14")
		padded := localParent.CreateElement("mpadded")
		padded.CreateAttr("height", "-0.5ex")
		padded.CreateAttr("depth", "0.5ex")
		padded.CreateAttr("voffset", "-0.5ex")
		row := padded.CreateElement("mrow")
		mi_e := row.CreateElement("mi")
		mi_e.SetText("E")
		space = localParent.CreateElement("mspace")
		space.CreateAttr("width", "-0.115em")
		mi_x := localParent.CreateElement("mi")
		mi_x.SetText("X")

		setFont(mi_t, mi_t.Tag, font)
		setFont(mi_e, mi_e.Tag, font)
		setFont(mi_x, mi_x.Tag, font)

	} else if strings.HasPrefix(token, OPERATORNAME) {

		element := parent.CreateElement("mo")
		addAttributes(element, attributes)
		element.SetText(token[14 : len(token)-1])

	} else if strings.HasPrefix(token, BACKSLASH) {

		element := parent.CreateElement("mi")
		addAttributes(element, attributes)

		if errCode == nil {
			element.SetText("&#x" + code + ";")
		} else if slices.Contains(FUNCTIONS, token) {
			element.SetText(token[1:])
		} else {
			element.SetText(token)
		}

		setFont(element, element.Tag, font)

	} else {
		element := parent.CreateElement("mi")
		addAttributes(element, attributes)
		element.SetText(token)
		setFont(element, element.Tag, font)
	}
}

func setFont(element *etree.Element, key string, font map[string]string) {
	if value, exist := font[key]; exist {
		element.CreateAttr("mathvariant", value)
	}
}
